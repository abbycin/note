/***********************************************
        File Name: trie
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 10/1/19 1:35 PM
***********************************************/

package routers

import (
	"strings"
)

type nodeType uint8

type IParam interface {
	set(k, v string)
	get(k string) string
}

type paramType struct {
	Key string
	Val string
}

type paramImpl []paramType

func (p *paramImpl) set(key, val string) {
	*p = append(*p, paramType{key, val})
}

func (p *paramImpl) get(key string) string {
	for _, item := range *p {
		if item.Key == key {
			return item.Val
		}
	}
	return ""
}

const (
	normal nodeType = iota
	root
	param
	wildcard
)

type IHandler interface {
	Serve(c *Context)
}

type node struct {
	path     string
	nType    nodeType
	wild     bool
	handles  []IHandler
	indices  []byte
	children []*node
}

func (n *node) addNode(path string, handle ...IHandler) {
	fullPath := path
	if len(n.path) > 0 || len(n.children) > 0 {
		n.append(path, fullPath, handle...)
	} else {
		// empty tree
		n.insert(path, fullPath, handle...)
		n.nType = root
	}
}

func (n *node) search(path string, params IParam) []IHandler {
travel:

	prefix := n.path
	if len(path) > len(prefix) {
		// check if has common prefix
		if path[:len(prefix)] == prefix {
			// if this node does not have a wildcard (param or wildcard)
			// child, we can just look up the next child node and continue
			// to travel down the tree
			path = path[len(prefix):]
			if !n.wild {
				c := path[0]
				indices := n.indices
				for i, max := 0, len(indices); i < max; i++ {
					if c == indices[i] {
						n = n.children[i]
						prefix = n.path
						goto travel
					}
				}

				return nil
			}

			// handles ':' and '*'
			n = n.children[0]
			switch n.nType {
			case param:
				end := 0
				for end < len(path) && path[end] != '/' {
					end++
				}

				// if the route is `/moha/:foo`, request path is '/moha/+1s'
				// here n.path is `:foo`
				params.set(n.path[1:], // so Key is foo
					path[:end]) // so Val is +1s

				// sub-path remains, go deeper
				if end < len(path) {
					if len(n.children) > 0 {
						path = path[end:]
						n = n.children[0]
						prefix = n.path
						goto travel
					}

					return nil
				}

				if handle := n.handles; handle != nil {
					return handle
				} else if len(n.children) == 1 {
					// check if a handles for this path + '/'
					// exists
					// FIXME: this may not be allowed, `/foo` an `/foo/` is different
					n = n.children[0]
					if n.path == "/" && n.handles != nil {
						return n.handles
					}
				}

				return nil

			case wildcard:
				// if route is `/foo/*mo`, request is `/foo/ha`
				// here n.path is `/*mo`
				params.set(n.path[2:], // skip `/*`
					path) // all reset path after substr prefix

				return n.handles

			default:
				panic("invalid node type")
			}
		}
	} else if path == prefix {
		handle := n.handles
		if handle != nil {
			return handle
		}

		// FIXME: this may not be allowed, `/foo` and `/foo/` is different
		for i, max := 0, len(n.indices); i < max; i++ {
			if n.indices[i] == '/' {
				n = n.children[i]
				return n.handles
			}
		}
	}
	return nil
}

func min(l, r int) int {
	if l < r {
		return l
	}
	return r
}

func (n *node) append(path, fullPath string, handles ...IHandler) {
travel:
	for {
		i := 0
		max := min(len(path), len(n.path))
		// find longest common prefix
		for i < max && path[i] == n.path[i] {
			i++
		}

		// split n.path into two parts
		// root take common prefix, child take reset part of path
		if i < len(n.path) {
			child := &node{
				path:     n.path[i:],
				wild:     n.wild,
				nType:    normal,
				indices:  n.indices,
				children: n.children,
				handles:  n.handles,
			}

			n.children = []*node{child}
			n.indices = []byte{n.path[i]}
			n.path = path[:i]
			n.handles = nil // reset
			n.wild = false
		}

		// make new node a child of this node
		if i < len(path) {
			path = path[i:]

			if n.wild {
				// since '/foo/:name' and '/foo/:name/xx'
				n = n.children[0]

				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					n.nType != wildcard &&
					(len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue travel

				} else {
					seg := path
					if n.nType != wildcard {
						seg = strings.SplitN(seg, "/", 2)[0]
					}
					prefix := fullPath[:strings.Index(fullPath, seg)] + n.path
					panic("'" + seg +
						"' in new path '" + fullPath +
						"'  conflicts with exsiting wildcard '" + n.path +
						"' in existing prefix '" + prefix + "'")
				}
			}

			c := path[0]

			// '/' after param
			if n.nType == param && c == '/' && len(n.children) == 1 {
				n = n.children[0]
				continue travel
			}

			// check if a child has prefix as c
			for i, max := 0, len(n.indices); i < max; i++ {
				if c == n.indices[i] {
					n = n.children[i]
					continue travel
				}
			}

			// otherwise insert it
			if c != ':' && c != '*' {
				n.indices = append(n.indices, c)
				child := &node{}
				n.children = append(n.children, child)
				n = child
			}

			n.insert(path, fullPath, handles...)

		} else if i == len(path) {
			if n.handles != nil && len(n.handles) != 0 {
				panic("a hanle is already registered for path " + fullPath)
			}
			n.handles = handles
		}
		break
	}
}

func (n *node) insert(path, fullPath string, handles ...IHandler) {
	var offset int = 0

	for i, max := 0, len(path); i < max; i++ {
		c := path[i]
		if c != ':' && c != '*' {
			continue
		}
		end := i + 1 // skip ':' or '*'

		for end < max && path[end] != '/' {
			if path[end] == ':' || path[end] == '*' {
				// eg, '/:ha:' or '/mo**' is invalid
				panic("only one placeholder per path is allowed")
			}
			end++
		}

		if len(n.children) > 0 {
			// eg, /foo/:mo and /foo/ha, or, /foo/mo* and /foo/ha is conflict
			panic("placeholder route conflict with existing one")
		}

		if end-i < 2 {
			panic("placeholder ':' or '*' must be named with a non-empty name")
		}

		if c == ':' {
			// split path at the beginning of the wildcard
			// eg, path is not start with :, like :name
			// or else, n.path remain empty
			if i > 0 {
				n.path = path[offset:i]
				offset = i
			}

			child := &node{
				nType: param,
			}
			n.children = []*node{child}
			n.wild = true
			n = child

			// if path not end with placeholder, then there will be another
			// non-placeholder sub-path starting with '/'
			// for example, path is '/foo/:name/:id'
			// above code may take '/foo/' or nothing
			// here may take ':name' or '/foo/:name'
			if end < max {
				n.path = path[offset:end]
				offset = end

				child := &node{}
				n.children = []*node{child}
				n = child
			}
		} else if c == '*' {

			if end != max {
				// eg, '/foo*/xx' is not allowed
				panic("wildcard must at the end of path")
			}

			if len(n.path) > 0 && n.path[len(n.path)-1] == '/' {
				// '/' is exist then wildcard is not allowed
				// eg, '/foo/' and '/foo/*' is conflict
				panic("wildcard conflicts with existing handles for the path root")
			}

			i--

			if path[i] != '/' {
				// check at first insert to an empty tree
				// eg, '/foo*xx' is not allowed
				panic("no / before wildcard path")
			}

			n.path = path[offset:i] // hold prefix

			child := &node{
				nType: wildcard,
				wild:  true,
			}
			n.children = []*node{child}
			n.indices = []byte{path[i]} // path[i] is '/'
			n = child

			child = &node{
				path:    path[i:], // `/*key`
				nType:   wildcard,
				handles: handles,
			}
			n.children = []*node{child}
			return
		}
	}
	// here may take '/:id'
	n.path = path[offset:]
	n.handles = handles
}
