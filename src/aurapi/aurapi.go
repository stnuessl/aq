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

package aurapi

import (
    "archive/tar"
    "bytes"
    "compress/gzip"
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "net/http"
    "sort"
)

var tag string = "AurAPI"

type AurAPI struct {
    client http.Client
    
    debug bool
}

type response struct {
    Type string
    Results interface{}
}

type PackageInfo struct {
    URL string
    Description string
    Version string
    Name string
    FirstSubmitted int
    License string
    ID int
    OutOfDate bool
    LastModified int
    Maintainer string
    CategoryID int
    URLPath string
    NumVotes int
}

type byNumVotes []PackageInfo

func (this byNumVotes) Len() int {
    return len(this)
}

func (this byNumVotes) Less(i, j int) bool {
    return this[j].NumVotes < this[i].NumVotes
}

func (this byNumVotes) Swap(i, j int) {
    this[i], this[j] = this[j], this[i]
}

func newPackageInfoFromJsonVal(val map[string]interface{}) *PackageInfo {
    ret := &PackageInfo{}
    
    url     := val["URL"]
    desc    := val["Description"]
    version := val["Version"]
    name    := val["Name"]
    sub     := val["FirstSubmitted"]
    license := val["License"]
    id      := val["ID"]
    ood     := val["OutOfDate"]
    last    := val["LastModified"]
    main    := val["Maintainer"]
    cat     := val["CategoryID"]
    path    := val["URLPath"]
    votes   := val["NumVotes"]
    
    if url != nil {
        ret.URL = url.(string)
    }
    
    if desc != nil {
        ret.Description = desc.(string)
    }
    
    if version != nil {
        ret.Version = version.(string)
    }
    
    if name != nil {
        ret.Name = name.(string)
    }
    
    if sub != nil {
        ret.FirstSubmitted = int(sub.(float64))
    }
    
    if license != nil {
        ret.License = license.(string)
    }
    
    if id != nil {
        ret.ID = int(id.(float64))
    }
    
    ret.OutOfDate = ood != nil
    
    if last != nil {
        ret.LastModified = int(last.(float64))
    }
    
    if main != nil {
        ret.Maintainer = main.(string)
    }
    
    if cat != nil {
        ret.CategoryID =  int(cat.(float64))
    }
    
    if path != nil {
        ret.URLPath = path.(string)
    }
    
    if votes != nil {
        ret.NumVotes = int(votes.(float64))
    }
    
    return ret
}

func (this PackageInfo) String() string {
    return fmt.Sprintf(
`Package [ %s ]:
  URL            : %s
  Description    : %s
  Version        : %s
  FirstSubmitted : %d
  License        : %s
  ID             : %d
  OutOfDate      : %t
  LastModified   : %d
  Maintainer     : %s
  CategoryID     : %d
  URLPath        : %s
  NumVotes       : %d`, this.Name, this.URL, 
                        this.Description, this.Version, this.FirstSubmitted,
                        this.License, this.ID, this.OutOfDate, 
                        this.LastModified, this.Maintainer, this.CategoryID,
                        this.URLPath, this.NumVotes)
}

func build_url(t string, arg string) string {
    buf := bytes.NewBufferString("https://aur.archlinux.org/rpc.php")
    
    buf.WriteString("?type=")
    buf.WriteString(t)
    buf.WriteString("&arg=")
    buf.WriteString(arg)
    
    return buf.String()
}

func invalid_response_type(t string, ex string) error {
    return fmt.Errorf("Invalid response type \"%s\" - expected \"%s\"\n", t, ex)
}

func  (this *AurAPI) getResponse(url string) (*response, error) {
    resp, err := this.request(url)
    
    ret := &response{}
    
    err = json.Unmarshal(resp, ret)
    if err != nil {
        return nil, err
    }
    
    if ret.Type == "error" {
        return nil, errors.New(ret.Results.(string))
    }
    
    return ret, nil
}

func (this *AurAPI) request(url string) ([]byte, error) {
    if this.debug {
        fmt.Printf("%s: GET: \"%s\"\n", tag, url)
    }
    
    msg, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil, err
    }
    
    resp, err := this.client.Do(msg)
    if err != nil {
        return nil, err
    }
    
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        return nil, errors.New(resp.Status)
    }
    
    ret, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    
    if this.debug {
        fmt.Printf("%s: response:\n%s\n", tag, string(ret))
    }
    
    return ret, nil
}

func NewAurAPI(debug bool) *AurAPI {
    return &AurAPI{http.Client{}, debug}
}

func (this *AurAPI) GetPackageInfo(pkg string) (*PackageInfo, error) {
    url := build_url("info", pkg)
    
    resp, err := this.getResponse(url)
    if err != nil {
        return nil, err
    }

    if resp.Type != "info" {
         return nil, invalid_response_type(resp.Type, "info")
    }
    
    m := resp.Results.(map[string]interface{})
    
    return newPackageInfoFromJsonVal(m), nil
}

func (this *AurAPI) SearchPackages(pkg string) ([]PackageInfo, error) {
    url := build_url("search", pkg)
    
    resp, err := this.getResponse(url)
    if err != nil {
        return nil, err
    }
    
    if resp.Type != "search" {
        return nil, invalid_response_type(resp.Type, "search")
    }
    
    list := resp.Results.([]interface{})
    ret := make([]PackageInfo, 0, len(list))
    
    for _, x := range list {
        info := newPackageInfoFromJsonVal(x.(map[string]interface{}))
        
        ret = append(ret, *info)
    }
    
    sort.Sort(byNumVotes(ret))
    
    return ret, nil
}

func getPackageUrl(urlpath string) string {
    buf := bytes.NewBufferString("https://aur.archlinux.org/")
    
    buf.WriteString(urlpath)
    
    return buf.String()
}

func (this *AurAPI) GetPackage(pkg string) (*tar.Reader, error) {
    info, err := this.GetPackageInfo(pkg)
    if err != nil {
        return nil, err
    }
    
    url := getPackageUrl(info.URLPath)
    
    buf, err := this.request(url)
    if err != nil {
        return nil, err
    }
    
    gzip, err := gzip.NewReader(bytes.NewReader(buf))
    if err != err {
        return nil, err
    }
    
    defer gzip.Close()
    
    return tar.NewReader(gzip), nil
}