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
        "fmt"
        "time"
        "flag"
        "io/ioutil"
        "crypto/md5"
        "encoding/binary"
        )

var ip = flag.String("ip", "", "Server IP")
var port = flag.Int("port", 0, "Server port")
var payload_path = flag.String("payload", "", "Path of the payload to be sent")
var runs = flag.Int("runs", 1, "Number of times the payload should be send")

func check(e error) {
    if e != nil {
        fmt.Fprintf(os.Stderr, "\n%s", e)
    }
    os.Exit(1)
}


func main() {

    flag.Parse()

    if *ip == "" || *port == 0 || *payload_path == "" {
        fmt.Println("Error while parsing command line arguments. Usage:")
        flag.PrintDefaults()
        os.Exit(1)
    }

    ipaddr := net.ParseIP(*ip)
    if ipaddr == nil {
        fmt.Fprintf(os.Stderr, "IP address is not valid")
        os.Exit(1)
    }

    fmt.Printf("Contacting server at %s:%d\n", ipaddr.String(), *port)
    tcpaddr := net.TCPAddr{
                IP:   ipaddr,
                Port: *port,
    }

    var retry int = 3
    var currRun int = *runs
    var conn net.Conn
    var err error


    for {
        conn, err = net.DialTimeout("tcp", tcpaddr.String(), time.Duration(2 * time.Second))
        if err != nil {
            retry -= 1
            if retry == 0 {
                fmt.Fprintf(os.Stderr, "Failed to connect\n")
                os.Exit(1)
            }
            fmt.Fprint(os.Stderr, "Connection failed, retrying...\n")
            time.Sleep(2 * time.Second)
            continue
        } else {
            break
        }
    }

    /* Connection succeeded, reset retry counter */
    retry = 3

    /* Reading payload, calculating md5sum and sending everything off to the client */
    data, err := ioutil.ReadFile(*payload_path)
    if(err != nil) {
        fmt.Fprintf(os.Stderr, "Could not read payload file: %s\n", err)
        conn.Close()
        os.Exit(1)
    }

    checksum := md5.Sum(data)
    signature := []byte{0xFD, 0xFD, 0xFD, 0xFD}

    for {
        if(currRun == 0) {
            conn.Close()
            break
        }

        /* Sending signature to the server */
        _, err = conn.Write(signature)
        if err != nil {
            fmt.Fprintf(os.Stderr, "Could not send signature to server: %s\n", err)
            conn.Close()
            os.Exit(1)
        }

        /* Sending the lenght of the payload, should move this to uint64 */
        err := binary.Write(conn, binary.LittleEndian, uint32(len(data)))
        if err != nil {
            fmt.Fprintf(os.Stderr, "Could not send payload lenght to client: %s\n", err)
            conn.Close()
            os.Exit(1)
        }

        _, err = conn.Write(checksum[:])
        if err != nil {
            fmt.Fprintf(os.Stderr, "Error while sending checksum: %s\n", err)
            conn.Close()
            os.Exit(1)
        }

        fmt.Fprintf(os.Stderr, "\rRun: %d", *runs - currRun + 1)
        currRun -= 1
    }
}

