package json

import jsoniter "github.com/json-iterator/go"

/*
 * @Descripttion:
 * @version:
 * @Author: dengdajun
 * @Date: 2020-12-23 16:45:44
 * @LastEditors: dengdajun
 * @LastEditTime: 2020-12-23 17:42:37
 */

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var Marshal = json.Marshal
var Unmarshal = json.Unmarshal
var NewDecoder = json.NewDecoder
