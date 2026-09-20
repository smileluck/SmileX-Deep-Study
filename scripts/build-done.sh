#!/usr/bin/env bash
# make build / make cross 的收尾提示。
# 为什么单独一个脚本：Scoop 版 GNU make 在 Windows 上按系统 ANSI 码页（GBK）转换
# 配方文本，Makefile 里直接写中文，经 make 传给子进程必乱码；放在本文件里由 bash
# 直接读取输出则不受影响。
# 用法：build-done.sh <产物路径> [hint]   带 hint 时追加本机运行提示
set -euo pipefail
printf '打包完成：%s\n' "$1"
if [ "${2:-}" = "hint" ]; then
  printf '运行后访问 http://127.0.0.1:5574；启动时自动铺出 AGENTS.md/.agents/.codebuddy 等工作区文件，非空目录自动收进 deepstudy/ 子目录，二进制留在原地\n'
fi
