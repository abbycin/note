#!/usr/bin/env bash

if [ $UID -ne 0 ]
then
  printf "\e[31msuperuser priviledge required\e[0m\n"
  exit 1
fi

DST=/opt/blog
mkdir $DST
cp -af assets $DST
cp -af blog_default.toml $DST/blog.toml
sed -i "s|ROOT|$DST|" $DST/blog.toml
cp -af blog.service /etc/systemd/system
sed -i "s|ROOT|$DST|g" /etc/systemd/system/blog.service
./build.sh
cp -af blog $DST

printf "\e[32mdone\e[0m\n"
