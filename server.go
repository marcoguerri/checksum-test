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
    "os"
    "fmt"
    "net"
    "flag"
    "log"
    "bytes"
    "io"
    "crypto/md5"
    "encoding/binary"
)

const headerlen = 4
const payloadlen = 4

var port = flag.Int("port", 8080, "Port to which the socket will be bound")
var ip = flag.String("ip", "", "IP address to which the socket will be bound")

func InitLogger(infoWriter io.Writer, errorWriter io.Writer) (*log.Logger, *log.Logger) {
    
    var infoLogger *log.Logger;
    var errorLogger *log.Logger;

    infoLogger = log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    errorLogger = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
    return infoLogger, errorLogger
}

func handleConnection(conn net.Conn) {

    var n int
    var err error
    var lengthInt int64

    header := make([]byte, headerlen)
    length := make([]byte, payloadlen)
    checksum := make([]byte, 16)
    
    for {
        /* Reads the header and compares it to the signature */
        header = make([]byte, headerlen)
        n, err = conn.Read(header)
        if err != nil || n < headerlen {
            conn.Close()
            return
        }

        if(binary.LittleEndian.Uint32(header) == 0xFDFDFDFD) {
            
            /* Reads the expected length of the payload */
            n, err = conn.Read(length)
            
            if err != nil || n < payloadlen {
                conn.Close()
                fmt.Fprintf(os.Stderr, "Could not read the expected lenght of the payload\n")
                return
            }

            n, err = conn.Read(checksum)
            if err != nil || n < 16 {
                conn.Close()
                return
            }

            /* Reads the payload itself */
            lengthInt =  int64(binary.LittleEndian.Uint32(length))
            fmt.Println(lengthInt)
            payload := make([]byte, lengthInt)

            n, err = conn.Read(payload)

            if (err != nil || n < int(lengthInt)) {
                conn.Close()
                fmt.Fprintf(os.Stderr, "Payload len (%d) shorter than expected (%d)\n",
                                        n, lengthInt)
                return
            }
            
            calculatedChecksum := md5.Sum(payload)
            if(bytes.Equal(calculatedChecksum[:], checksum)) {
                fmt.Fprintf(os.Stderr, "Checksums do not match\n")
            }  else {
                fmt.Fprintf(os.Stderr, "Checksum ok from %s\n", conn.RemoteAddr())
            }
        } else {
            fmt.Fprintf(os.Stderr, "Signature sent from the server not reconized\n")
            conn.Close()
            return
        }
    }
}

func main () {
    
    infoLogger, errorLogger := InitLogger(os.Stderr, os.Stderr)
    flag.Parse()
    var bind bytes.Buffer
    
    bind.WriteString(fmt.Sprintf("%s:%d", *ip, *port))
    listener, err := net.Listen("tcp", bind.String())
    if(err != nil) {
        errorLogger.Println("Error while binding socket to port", *port)
        errorLogger.Println(err)
        os.Exit(1)
    }

    /*
     * GO philosophy is to to provide blocking interfaces that are used
     * concurrently with goroutines and channels rather than polling with
     * callbacks. The obvious consequence is that the routine which handles
     * the I/O is relatively easy to write. Underneath, GO uses the async 
     * interfaces provided by the OS (epoll) to be notified for network events
     */
    
    for {
        conn ,err := listener.Accept()
        if err != nil {
            fmt.Fprintf(os.Stderr, "Connection closed unexpectedly\n")
            continue
        }

        infoLogger.Println("Accepted connection with ", conn.RemoteAddr())
        go handleConnection(conn)
   }
}
