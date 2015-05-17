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

package progopts

import (
    "fmt"
    "strconv"
)

type ProgOpts struct {
    opts []option
}

type option struct {
    sOpt string
    lOpt string
    val interface{}
    desc string
}

type optionMap map[string]option

func New() *ProgOpts {
    opts := make([]option, 0, 16)
    
    return &ProgOpts{opts}
}

func (this *ProgOpts) Add(s string, l string, val interface{}, desc string) {
    this.opts = append(this.opts, option{s, l, val, desc})
}

func expected1ArgErr(opt string) error {
    msg := "Option \"%s\" expects one argument but zero were given"
    return fmt.Errorf(msg, opt)
}

func expectedAtLeastOneArg(opt string) error {
    return fmt.Errorf("Option \"%s\" expects at least one argument", opt)
}

func handle(arg string, i int, optMap optionMap, args []string) (int, error) {
    opt, ok := optMap[arg]
    if ok {
        j := i + 1;
        
        /* get possible args */
        for ; j < len(args); j++ {
            _, ok = optMap[args[j]]
            if ok {
                break
            }
        }
        
        j--
        
        switch val := opt.val.(type) {
        case *bool:

            *val = true
        case *string:
            if j - i < 1 {
                return 0, expected1ArgErr(opt.lOpt)
            }
            
            j = i + 1
            
            *val = args[j]
        case *[]string:
            if j - i == 0 {
                return 0, expectedAtLeastOneArg(opt.lOpt)
            }
            
            *val = args[i + 1:j + 1]
        case *int:
            if j - i < 1 {
                return 0, expected1ArgErr(opt.lOpt)
            }
            
            j = i + 1
            
            tmp, err := strconv.Atoi(args[j])
            if err != nil {
                return 0, err
            }
            
            *val = tmp
        case *[]int:
            if j - i == 0 {
                return 0, expectedAtLeastOneArg(opt.lOpt)
            }
            
            for _, x := range args[i + 1:j + 1] {
                tmp, err := strconv.Atoi(x)
                if err != nil {
                    return 0, err
                }
                
                *val = append(*val, tmp)
            }
        default:
            return 0, fmt.Errorf("Invalid value type")
        }
        
        return j, nil
    } else {
        return 0, fmt.Errorf("unknown option \"%s\"", arg)
    }
}

func (this *ProgOpts) ParseArgs(args []string) error {
    optMap := make(optionMap, len(this.opts))
    
    for _, x := range this.opts {
        if len(x.sOpt) > 0 {
            optMap["-" + x.sOpt] = x
        }
        
        if len(x.lOpt) > 0 {
            optMap["--" + x.lOpt] = x
        }
    }
    
    for i := 0; i < len(args); i++ {
        if len(args[i]) < 2 || args[i][0] != '-' {
            return fmt.Errorf("expected option, got \"%s\"", args[i])
        }
        
        if args[i][0] == '-' && args[i][1] != '-' {
            for _, x := range args[i][1:] {
                j, err := handle("-" + string(x), i, optMap, args)
                if err != nil {
                    return err
                }
                
                i = j
            }
        } else {
            j, err := handle(args[i], i, optMap, args)
            if err != nil {
                return err
            }
            
            i = j
        }
    }
    
    return nil
}

func (this *ProgOpts) Usage(usage string) {
    fmt.Printf("Usage: %s\n", usage)
    
    for _, x := range this.opts {
        if len(x.sOpt) > 0 && len(x.lOpt) > 0 {
            fmt.Printf("  -%s  --%-16s %s\n", x.sOpt, x.lOpt, x.desc)
        } else if len(x.sOpt) > 0 {
            fmt.Printf("  -%s    %-16s %s\n", x.sOpt, "", x.desc)
        } else if len(x.lOpt) > 0 {
            fmt.Printf("  %s  --%-16s %s\n", "  ", x.lOpt, x.desc)
        }
    }
}