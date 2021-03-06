package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/subscribers/protocol"
	"github.com/naveego/navigator-go/subscribers/server"
	"github.com/sirupsen/logrus"
)

func main() {

	logrus.SetOutput(os.Stdout)

	if len(os.Args) < 2 {
		fmt.Println("Not enough arguments.")
		os.Exit(-1)
	}

	flag.Parse()

	addr := os.Args[1]

	logrus.SetLevel(logrus.DebugLevel)

	logrus.WithField("listen-addr", addr).Info("Started console_subscriber")

	srv := server.NewSubscriberServer(addr, &subscriberHandler{})

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatal("Error shutting down server: ", err)
		}
	}()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, os.Kill)
	fmt.Println("CTRL-C to close")

	<-signals

	fmt.Println("Shutting down.")
}

type subscriberHandler struct {
	fileWriter io.WriteCloser
}

func (h *subscriberHandler) Init(request protocol.InitRequest) (protocol.InitResponse, error) {
	logrus.Debugf("Init: %#v", request)

	if request.Settings != nil {
		if fileName, ok := request.Settings["file"]; ok && fileName != "" {
			f, err := os.Create(fileName.(string))
			if err != nil {
				return protocol.InitResponse{
					Success: false,
					Message: "couldn't open file: " + err.Error(),
				}, err
			}
			h.fileWriter = f
		}

	}

	return protocol.InitResponse{
		Success: true,
		Message: "OK",
	}, nil
}

func (h *subscriberHandler) TestConnection(request protocol.TestConnectionRequest) (protocol.TestConnectionResponse, error) {
	logrus.Debugf("TestConnection: %#v", request)

	return protocol.TestConnectionResponse{
		Success: true,
		Message: "OK",
	}, nil
}

func (h *subscriberHandler) DiscoverShapes(request protocol.DiscoverShapesRequest) (protocol.DiscoverShapesResponse, error) {
	logrus.Debugf("DiscoverShapes: %#v", request)

	return protocol.DiscoverShapesResponse{
		Shapes: pipeline.ShapeDefinitions{
			pipeline.ShapeDefinition{
				Name:        "test-shape",
				Description: "test-shape description",
				Keys:        []string{"ID"},
				Properties: []pipeline.PropertyDefinition{
					{
						Name: "ID",
						Type: "number",
					},
					{
						Name: "Name",
						Type: "string",
					},
				},
			},
		},
	}, nil
}

func (h *subscriberHandler) ReceiveDataPoint(request protocol.ReceiveShapeRequest) (protocol.ReceiveShapeResponse, error) {
	logrus.WithField("datapoint", request.DataPoint).Info(color(42, "Received DataPoint"))

	if h.fileWriter != nil {
		jsonBytes, _ := json.Marshal(request.DataPoint)
		fmt.Fprintln(h.fileWriter, string(jsonBytes))
	}

	return protocol.ReceiveShapeResponse{
		Success: true,
	}, nil
}

func (h *subscriberHandler) Dispose(request protocol.DisposeRequest) (protocol.DisposeResponse, error) {
	logrus.Debugf("Dispose: %#v", request)

	if h.fileWriter != nil {
		_ = h.fileWriter.Close()
	}

	return protocol.DisposeResponse{
		Success: true,
	}, nil
}

func color(code int, s string) string {
	return fmt.Sprintf("\033[%dm%s\033[0m", code, s)
}
