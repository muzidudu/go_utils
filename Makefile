# go_utils 根目录 Makefile - 委托给 fiber
# 用法: make [target]  等同于 make -C fiber [target]

.PHONY: help build run deploy clean sync test lint

help:
	@make -C fiber help

build run deploy clean test lint:
	@make -C fiber $@

sync:
	@go work sync
