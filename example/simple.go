// An example of using the simple go-remctl client library
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

    flag.StringVar( &princ, "s", "", "remctl service principal (default host/<hostname>)" )
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

    result, err := remctl.Remctl( host, port, princ, args[1:] )

    if err != nil {
	fmt.Println( err )
    } else {
	fmt.Fprintf( os.Stdout, result.Stdout )
	fmt.Fprintf( os.Stderr, result.Stderr )
	os.Exit( result.Status )
    }

}
