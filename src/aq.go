/*
 * Copyright (C) 2015  Steffen Nüssle
 * aq - aur query
 *
 * This file is part of aq.
 *
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
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
    "archive/tar"
    "fmt"
    "io"
    "os"
    "log"
    "strings"
)

import (
    "aurapi"
    "progopts"
)

func untar(tarReader *tar.Reader) error {
    fmt.Printf("Unzipping ... - ")
    
    for {
        header, err := tarReader.Next()
        if err != nil {
            if err == io.EOF {
                break
            } else {
                return err
            }
        }
        
        finfo := header.FileInfo()
        
        if finfo.IsDir() {
            err = os.Mkdir(header.Name, 0755)
            if err != nil {
                if !os.IsExist(err) {
                    return err
                }
            }
        } else {
            flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
            
            file, err := os.OpenFile(header.Name, flags, 0644)
            if err != nil {
                return err
            }
                    
            _, err = io.Copy(file, tarReader)
            if err != nil {
                return err
            }
        }
    }
    
    fmt.Printf("done.\n")
    
    return nil
}

func main() {
    pkgs    := make([]string, 0, 10)
    info    := make([]string, 0, 10)
    search  := make([]string, 0, 10)
    limit   := 10
    help    := false
    debug   := false
    
    opts := progopts.New()

    opts.Add("p", "package",    &pkgs,      "Download and unpack a package.")
    opts.Add("i", "info",       &info,      "Show package information.")
    opts.Add("s", "search",     &search,    "Search for a package.")
    opts.Add("l", "limit",      &limit,     "Limit amount of search results.")
    opts.Add("",  "help",       &help,      "Show this help message.")
    opts.Add("",  "debug",      &debug,     "Enable debug output.")
    
    err := opts.ParseArgs(os.Args[1:])
    if err != nil {
        log.Fatal(err)
    }
    
    if help {
        opts.Usage(os.Args[0] + " --opt1 [arg1] [arg2] --opt2 [arg1] ...")
        os.Exit(0)
    }

    api := aurapi.NewAurAPI(debug)
    
    for _, x := range info {
        pkgInfo, err := api.PackageInfo(x)
        if err != nil {
            log.Fatal(err)
        }
                
        fmt.Printf("%s\n", pkgInfo)
    }
        
    for _, x := range search {
        pkgInfoList, err := api.Search(x, limit)
        if err != nil {
            log.Fatal(err)
        }
        
        line := strings.Repeat("-", 50)
        
        fmt.Printf("%7s  %-30s  %s\n%s\n", "Votes", "Package Name", "Id", line)
        
        for _, y := range pkgInfoList {
            fmt.Printf("%7d  %-30s  %7d\n", y.NumVotes, y.Name, y.ID)
        }
    }
    
    for _, x := range pkgs {
        tarReader, err := api.Package(x)
        if err != nil {
            log.Fatal(err)
        }
        
        err = untar(tarReader)
        if err != nil {
            log.Fatal(err)
        }
    }
}