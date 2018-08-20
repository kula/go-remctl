// An example of using the complex go-remctl client library
//
// Copyright 2014 Thomas L. Kula <kula@tproa.net>
// All Rights Reserved
//
// Use of this source code is governed by a BSD-style license that can 
// be found in the LICENSE file.

package main

import (
    "flag"
    "fmt"
    "os"
    "strings"
    remctl "github.com/kula/go-remctl/client"
)


func main() {

    var host, princ string
    var port uint16
    var timeout uint

    flag.StringVar( &princ, "s", "", "remctl service principal (default host/<hostname>)" )
    flag.UintVar(&timeout, "t", 0, "timeout in seconds. Defaults to zero")
    flag.Parse()

    args := flag.Args()

    if len( args ) < 2 {
	fmt.Printf( "Usage: %s [options] <host> <command> [<args>...]\n", os.Args[0] )
	os.Exit( 1 )
    }

    // args[0] is the host, which can be host or host:port

    hostparts := strings.SplitN( args[0], ":", 2 )
    if len( hostparts ) == 1 {
	port = 0;
    } else {
	_, err := fmt.Sscanf( hostparts[1], "%d", &port )
	if err != nil {
	    fmt.Println( err )
	    os.Exit( 1 )
	}
    }

    if hostparts[0] == "" {
	fmt.Println( "Error: must specify host" )
	os.Exit( 1 )
    } else {
	host = hostparts[0]
    }

    remc, err := remctl.New()
    if err != nil {
	fmt.Println( err )
	os.Exit( 1 )
    }

	if timeout > 0 {
		if err := remc.SetTimeout(timeout); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

    if err := remc.Open( host, port, princ ); err != nil {
	fmt.Println( err )
	os.Exit( 1 )
    }

    if err := remc.Execute( args[1:] ); err != nil {
	fmt.Println( err )
	os.Exit( 1 )
    }

    exit := 0
    done := false

    for {
	select {
	case out := <-remc.Output:
	    // default to stderr
	    stream := os.Stderr
	    if out.Stream == 0 {
		stream = os.Stdout
	    }
	    fmt.Fprintf( stream, out.Data )
	case status := <-remc.Status:
	    exit = int(status)
	    done = true
	    break
	case err := <-remc.Error:
	    fmt.Fprintf( os.Stderr, err.Data )
	    done = true
	    break
	case <-remc.Done:
	    done = true
	    break
	}

	if done {
	    break
	}
    }

    remc.Close()
    os.Exit( exit )

}
