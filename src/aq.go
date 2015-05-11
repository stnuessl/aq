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
)

import (
    "aurapi"
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
    
    pkgName := os.Args[1]
    
    api := aurapi.NewAurAPI(true)
    
    tarReader, err := api.GetPackage(pkgName)
    if err != nil {
        log.Fatal(err)
    }
    
    err = untar(tarReader)
    if err != nil {
        log.Fatal(err)
    }
}