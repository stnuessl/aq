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