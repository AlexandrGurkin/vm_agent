package main

import (
	"fmt"
	"net"
	"os"
	"time"

	"github.com/AlexandrGurkin/common/consts"
	"github.com/AlexandrGurkin/common/middlewares"
	"github.com/AlexandrGurkin/common/xlog"
	"github.com/AlexandrGurkin/common/xlog/xzerolog"
	"github.com/AlexandrGurkin/tasker/client/api"
	"github.com/AlexandrGurkin/tasker/client/api/exec_task"
	"github.com/AlexandrGurkin/tasker/client/api/reg_agent"
	"github.com/AlexandrGurkin/tasker/client/api/set_result"
	"github.com/AlexandrGurkin/tasker/client/models"
	"github.com/AlexandrGurkin/tasker/restapi"
	verApi "github.com/AlexandrGurkin/vm_agent/client/api"
	"github.com/AlexandrGurkin/vm_agent/client/api/version"
	"github.com/AlexandrGurkin/vm_agent/handlers"
	rapi "github.com/AlexandrGurkin/vm_agent/restapi"
	"github.com/AlexandrGurkin/vm_agent/restapi/operations"
	"github.com/go-openapi/loads"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

func main() {
	logger := xzerolog.NewXZerolog(xlog.LoggerCfg{
		Level: "trace",
		Out:   os.Stdout,
	})

	httptransport.DefaultTimeout = time.Duration(100) * time.Second

	transport := httptransport.New("188.187.1.101:8077", "/api/v1", []string{"http"})
	transport.Debug = false

	taskerApi := api.New(transport, strfmt.Default)

	go func() {
		for {
			addrs := networks()
			_, err := taskerApi.RegAgent.PostAgent(reg_agent.NewPostAgentParams().WithBody(&models.RequestAgent{
				ID:   hostName(),
				Nets: addrs,
			}))
			if err != nil {
				logger.Errorf("agent reg with error [%s]", err.Error())
			} else {
				logger.Tracef("send addresses %v", addrs)
			}
			time.Sleep(time.Second)
		}
	}()

	go func() {
		for {
			agentID := hostName()
			res, err := taskerApi.ExecTask.GetExecuteTaskAgent(exec_task.NewGetExecuteTaskAgentParams().WithAgent(agentID))
			if err != nil {
				logger.Errorf("error get task for agent [%s] [%s]", agentID, err.Error())
				time.Sleep(time.Second * 10)
				continue
			}
			arg := res.GetPayload()
			logger.Infof("for agent [%s], request [%s], with command [%s] and target [%s]", arg.Agent, arg.ID, arg.Command, arg.Target)

			resp := ""

			switch arg.Command {
			case "connect":
				{
					cn, err := net.Dial("tcp", arg.Target)
					if err != nil {
						resp = err.Error()
					} else {
						_ = cn.Close()
						resp = "ok"
					}

				}
			case "ping":
				{
					resp = "oki"
				}
			case "version":
				transport := httptransport.New(arg.Target, "/api/v1", []string{"http"})
				transport.Debug = false

				vApi := verApi.New(transport, strfmt.Default)
				_, err := vApi.Version.GetVersion(version.NewGetVersionParams())
				if err != nil {
					resp = err.Error()
				} else {
					resp = "ok"
				}

			default:
				resp = "oki"
			}

			_, err = taskerApi.SetResult.PostSetResultID(set_result.NewPostSetResultIDParams().WithID(arg.ID).WithBody(&models.ResponseTask{
				ID:     arg.ID,
				Status: resp,
			}))
			if err != nil {
				logger.Errorf("error set result for request [%s] [%s]", arg.ID, err.Error())
				continue
			}
		}
	}()

	var err error
	swaggerSpec, err := loads.Embedded(restapi.SwaggerJSON, restapi.FlatSwaggerJSON)
	if err != nil {
		logger.Fatal(err.Error())
	}

	api := operations.NewTemplateForHTTPServerAPI(swaggerSpec)
	server := rapi.NewServer(api)

	server.Host = "0.0.0.0"
	server.Port = 8071
	rapi.SetMiddlewareConfig(middlewares.MiddlewareConfig{Logger: logger, Pprof: true})
	api.Logger = func(s string, i ...interface{}) {
		logger.WithXFields(xlog.Fields{consts.FieldModule: "swagger_api_logger"}).
			Infof(s, i...)
	}
	api.VersionGetVersionHandler = handlers.VersionHandler{}
	server.ConfigureAPI()
	server.KeepAlive = 10 * time.Second

	if err = server.Serve(); err != nil {
		logger.Fatal(err.Error())
	}

}

func hostName() string {
	hostname, err := os.Hostname()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return hostname
}

func networks() []string {
	res := []string{}
	ifaces, err := net.Interfaces()
	if err != nil {
		fmt.Println(err)
		return res
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			fmt.Println(err)
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip.IsLoopback() {
				continue
			}
			ok := ip.To4()
			if ok != nil {
				res = append(res, ip.String())
			}
		}
	}
	return res
}
