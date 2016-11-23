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

const signaturelen  = 4
const payloadlen    = 4
const checksumlen   = 16

var port = flag.Int("port", 8080, "Port to which the socket will be bound")
var ip = flag.String("ip", "", "IP address to which the socket will be bound")

func InitLog(infoWriter io.Writer, errorWriter io.Writer) (*log.Logger, *log.Logger) {

    var infoLogger *log.Logger;
    var errorLogger *log.Logger;
    infoLogger = log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
    errorLogger = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
    return infoLogger, errorLogger
}

func handleConnection(conn net.Conn, errorLogger *log.Logger) {

    header := make([]byte, signaturelen)
    length := make([]byte, payloadlen)
    checksum := make([]byte, 16)

    for {
        /* Reads the header and compares it to the signature */
        header = make([]byte, signaturelen)
        _, err := io.ReadFull(conn, header)
        if err != nil {
            errorLogger.Println("Error while reading signature", err)
            conn.Close()
            return
        }

        if(binary.LittleEndian.Uint32(header) != 0xFDFDFDFD) {
            continue
        }

        /* Reading payload lenght */
        _, err = io.ReadFull(conn, length)
        if err != nil {
            errorLogger.Println("Error while reading payload lenght", err)
            conn.Close()
            return
        }

        /* Reading checksum */
        _, err = io.ReadFull(conn, checksum)
        if err != nil {
            errorLogger.Println("Error while reading the checksum", err)
            conn.Close()
            return
        }

        /* Reading payload */
        payload := make([]byte, binary.LittleEndian.Uint32(length))
        _, err = io.ReadFull(conn, payload)
        if err != nil {
            if err == io.EOF {
                /* Connection was closed by remote host */
                return
            } else {
                errorLogger.Println("Error while reading payload", err)
                conn.Close()
                return
            }
       }

       calculatedChecksum := md5.Sum(payload)
       if(bytes.Equal(calculatedChecksum[:], checksum)) {
           errorLogger.Println("Checksums do not match!")
       }
    }
}

func main () {

    infoLogger, errorLogger := InitLog(os.Stderr, os.Stderr)
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
        go handleConnection(conn, errorLogger)
   }
}
