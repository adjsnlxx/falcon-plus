// Copyright 2017 Xiaomi, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	//"github.com/open-falcon/falcon-plus/modules/agent/cron"
	"../agent/cron"
	//"github.com/open-falcon/falcon-plus/modules/agent/funcs"
	"../agent/funcs"
	//"github.com/open-falcon/falcon-plus/modules/agent/g"
	"../agent/g"
	//"github.com/open-falcon/falcon-plus/modules/agent/http"
	"../agent/http"
	"os"
)

func main() {

	g.BinaryName = BinaryName
	g.Version = Version
	g.GitCommit = GitCommit

	// 命令行参数
	cfg := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	check := flag.Bool("check", false, "check collector")

	// 获取一下命令行参数
	flag.Parse()

	// 打印版本
	if *version {
		fmt.Printf("Open-Falcon %s version %s, build %s\n", BinaryName, Version, GitCommit)
		os.Exit(0)
	}

	// 检查是否能正常采集监控数据
	if *check {
		funcs.CheckCollector()
		os.Exit(0)
	}

	// 读取配置文件
	g.ParseConfig(*cfg)

	// 设置日志级别
	if g.Config().Debug {
		g.InitLog("debug")
	} else {
		g.InitLog("info")
	}

	// 初始化
	g.InitRootDir() // 初始化根目录
	g.InitLocalIp() // 初始化本地 IP
	g.InitRpcClients()// 初始化 hbs rpc 客户端

	// 建立收集数据的函数和发送数据间隔映射
	funcs.BuildMappers()

	// 定时任务（每秒），通过文件 /proc/cpustats 和 /proc/diskstats，更新 cpu 和 disk 的原始数据，等待后一步处理
	go cron.InitDataHistory()

	// 定时任务
	cron.ReportAgentStatus()// 向 hbs 发送客户端心跳信息
	cron.SyncMinePlugins()// 同步 plugins
	cron.SyncBuiltinMetrics()// 从 hbs 同步 BuiltinMetrics
	cron.SyncTrustableIps()// 同步可信任 IP 地址
	cron.Collect()// 采集监控数据

	// 开启 http 服务
	go http.Start()

	// 阻塞 main 函数
	select {}

}
