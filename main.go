/***********************************************
        File Name: routers
        Author: Abby Cin
        Mail: abbytsing@gmail.com
        Created Time: 9/14/19 3:50 PM
***********************************************/

package main

import (
	"blog/conf"
	"blog/dbutil"
	"blog/logging"
	"blog/middleware"
	"blog/routers"
	"blog/service"
	"blog/session"
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

var (
	help    bool
	cfgPath string
)

func init() {
	flag.BoolVar(&help, "h", false, "show help")
	flag.StringVar(&cfgPath, "conf", "", "config file path")
}

func main() {
	flag.Parse()
	if help || cfgPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	cfg := conf.Init(cfgPath)
	if err := logging.Init(cfg.GetLogger()); err != nil {
		fmt.Printf("can't init logger: %s\n", err)
		os.Exit(1)
	}
	defer logging.Release()

	mgr := session.NewSessionManager(cfg)
	mgr.StartGC()

	access := middleware.NewAccess()

	r := routers.New(mgr, cfg.Common.ProxyMode)
	r.ServeFile(cfg.Service.Assets, cfg.Common.Assets)
	r.AddCb(access)

	dao := dbutil.NewDao(cfg.Common.DbFile)

	auth := middleware.NewAuth(mgr, cfg)

	service.NewImage(cfg, r, auth)
	service.NewLogin(cfg, dao, r)
	service.NewManage(cfg, dao, r, auth)
	service.NewUser(cfg, dao, r, auth)
	post := service.NewArticle(cfg, dao, r)
	home := service.NewHome(cfg, dao, r)
	navi := service.NewNavi(cfg, dao, r, auth)
	edit := service.NewEditor(cfg, dao, r, auth)

	navi.Add(post)
	navi.Add(home)
	edit.Add(post)
	edit.Add(home)

	server := &http.Server{
		Addr:    cfg.Common.Addr,
		Handler: r,
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for {
			s := <-sig
			logging.Info("get signal %s", s.String())
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP:
				// must shutdown http server first
				err := server.Shutdown(context.Background())
				if err != nil {
					logging.Error("http server shutdown: %s", err)
				}
				dao.Close()
				logging.Info("server exited")
				return
			default:
				return
			}
		}
	}()

	err := server.ListenAndServe()
	if err != nil {
		logging.Fatal("ListenAndServe: %v", err)
	}
}
