/*
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>
 *
 * Author: Marco Guerri <marco.guerri.dev@fastmail.com>
 */
package main

import (
        "net"
        "os"
        "time"
        "log"
        "flag"
        "io"
        "io/ioutil"
        "crypto/md5"
        "encoding/binary"
        )

const timeout = 5

var ip = flag.String("ip", "", "Server IP")
var port = flag.Int("port", 0, "Server port")
var payload_path = flag.String("payload", "", "Path of the payload to be sent")
var runs = flag.Int("runs", 1, "Number of times the payload should be send")

func InitLog(infoWriter io.Writer, errorWriter io.Writer) (*log.Logger, *log.Logger) {
    var infoLogger *log.Logger;
    var errorLogger *log.Logger;
    infoLogger = log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    errorLogger = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
    return infoLogger, errorLogger
}

func main() {
   
    infoLogger, errorLogger := InitLog(os.Stderr, os.Stderr)
    flag.Parse()

    _, err := os.Stat(*payload_path)
    if os.IsNotExist(err) {
        errorLogger.Println("Payload does not exist ")
        os.Exit(1)
    }

    ipaddr := net.ParseIP(*ip)
    if ipaddr == nil {
        errorLogger.Println("IP address is not valid")
        os.Exit(1)
    }

    infoLogger.Println("Connecting server at", ipaddr.String(), ":",  *port)
    tcpaddr := net.TCPAddr{
                IP:   ipaddr,
                Port: *port,
    }

    var retry int = 3
    var conn net.Conn

    for {
        conn, err = net.DialTimeout("tcp", tcpaddr.String(), time.Duration(timeout * time.Second))
        if err != nil {
            retry -= 1
            if retry == 0 {
                errorLogger.Println("Failed to connect to server")
                os.Exit(1)
            }
            infoLogger.Println("Connection attempt failed, retrying in", timeout, "s")
            time.Sleep(timeout * time.Second)
            continue
        } else {
            retry = 3
            break
        }
    }

    data, err := ioutil.ReadFile(*payload_path)
    if(err != nil) {
        errorLogger.Println("Could not read payload file: %s", err)
        conn.Close()
        os.Exit(1)
    }

    checksum := md5.Sum(data)
    signature := []byte{0xFD, 0xFD, 0xFD, 0xFD}

    for {
        if(*runs == 0) {
            conn.Close()
            break
        }

        _, err = conn.Write(signature)
        if err != nil {
            errorLogger.Println("Could not send signature to server: %s", err)
            conn.Close()
            os.Exit(1)
        }

        err := binary.Write(conn, binary.LittleEndian, uint32(len(data)))
        if err != nil {
            errorLogger.Println("Could not send payload lenght to client:", err)
            conn.Close()
            os.Exit(1)
        }

        _, err = conn.Write(checksum[:])
        if err != nil {
            errorLogger.Println("Error while sending checksum:", err)
            conn.Close()
            os.Exit(1)
        }

        _, err = conn.Write(data)
        if err != nil {
            errorLogger.Println("Could not send payload to client", err)
            conn.Close()
            os.Exit(1)
        }

        *runs -= 1
        infoLogger.Println("Remaining runs:", *runs)
    }
}

