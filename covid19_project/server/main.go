package main

//Required Packages
import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"strings"

	DataFile "github.com/inam700/GoLang-Projects/covid19_project/client"
)

//Covid19 csv file
var dataset = DataFile.Load("../resources/covid_final_data.csv")

func main() {
	var address string
	var network string

	flag.StringVar(&address, "e", ":4040", "service endpoint [ip address or socket path]")
	flag.StringVar(&network, "n", "tcp", "network protocol [tcp,unix]")

	flag.Parse()

	switch network {
	//tcp
	case "tcp", "tcp4", "tcp6", "unix":
	default:
		log.Fatalln("unsupported network protocol: ", network)
	}
	ln, err := net.Listen(network, address)

	if err != nil {
		log.Fatal("failed to create listener", err)
	}
	defer ln.Close()
	log.Println("*** COVID19 PROJECT ***")
	log.Printf("Service Started: (%s) %s \n", network, address)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			if err := conn.Close(); err != nil {
				log.Println("failed to close listner: ", err)

			}
			continue

		}
		log.Println("Connected to ", conn.RemoteAddr())
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			log.Println("error closing connection", err)
		}
	}()
	if _, err := conn.Write([]byte(
		"Connected with the server.....\n" +
			"Now to get result you need to give query.....\n" +
			"first write get and then write region or date.....\n",
	)); err != nil {
		log.Println("error Writing", err)
		return
	}

	for {
		cmdLine := make([]byte, (1024 * 4))
		n, err := conn.Read(cmdLine)
		if n == 0 || err != nil {
			log.Println("Connection Read Error", err)
			return
		}
		cmd, param := parseCommand(string(cmdLine[0:n]))
		if cmd == "" {
			if _, err := conn.Write([]byte("Invalid  \n")); err != nil {
				log.Println("failed to write", err)
				return
			}
			continue
		}

		switch strings.ToUpper(cmd) {
		case "GET":
			result := DataFile.Find(dataset, param)
			if len(result) == 0 {
				if _, err := conn.Write([]byte("Nothing Found \n")); err != nil {
					log.Println("failed to write", err)
				}
				continue
			}
			covidJson, _ := json.MarshalIndent(result, "", " ")

			_, err := conn.Write([]byte(
				fmt.Sprintf(string(covidJson)),
			))
			if err != nil {
				log.Println("failed to write response", err)
				return
			}

		default:
			if _, err := conn.Write([]byte("Invalid Command \n")); err != nil {
				log.Println("Failed", err)
				return
			}
		}
	}
}

func parseCommand(cmdLine string) (cmd, param string) {
	parts := strings.Split(cmdLine, " ")
	if len(parts) != 2 {
		return "", ""
	}
	cmd = strings.TrimSpace(parts[0])
	param = strings.TrimSpace(parts[1])
	return
}
