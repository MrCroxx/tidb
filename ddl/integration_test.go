// Copyright 2021 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ddl_test

import (
	"fmt"
	"testing"

	"github.com/pingcap/tidb/testkit"
	"github.com/pingcap/tidb/util/collate"
)

func TestDefaultValueIsBinaryString(t *testing.T) {
	collate.SetCharsetFeatEnabledForTest(true)
	defer collate.SetCharsetFeatEnabledForTest(false)
	store, clean := testkit.CreateMockStore(t)
	defer clean()
	tests := []struct {
		colTp  string
		defVal string
		result string
	}{
		{"char(10) charset gbk", "0xC4E3BAC3", "你好"},
		{"char(10) charset gbk", "'好'", "好"},
		{"varchar(10) charset gbk", "0xC4E3BAC3", "你好"},
		{"char(10) charset utf8mb4", "0xE4BDA0E5A5BD", "你好"},
		{"char(10) charset utf8mb4", "0b111001001011100010010110111001111001010110001100", "世界"},
		{"bit(48)", "0xE4BDA0E5A5BD", "你好"},
		{"enum('你好')", "0xE4BDA0E5A5BD", "你好"},
		{"set('你好')", "0xE4BDA0E5A5BD", "你好"},
	}
	tk := testkit.NewTestKit(t, store)
	tk.MustExec("use test;")
	for _, tt := range tests {
		tk.MustExec("drop table if exists t;")
		template := "create table t (a %s default %s);"
		tk.MustExec(fmt.Sprintf(template, tt.colTp, tt.defVal))
		tk.MustExec("insert into t values (default);")
		tk.MustQuery("select a from t;").Check(testkit.Rows(tt.result))
	}

	// Test invalid default value.
	tk.MustExec("drop table if exists t;")
	// 0xE4BDA0E5A5BD81 is an invalid utf-8 string.
	tk.MustGetErrMsg("create table t (a char(20) charset utf8mb4 default 0xE4BDA0E5A5BD81);",
		"[ddl:1067]Invalid default value for 'a'")
	tk.MustGetErrMsg("create table t (a blob default 0xE4BDA0E5A5BD81);",
		"[ddl:1101]BLOB/TEXT/JSON column 'a' can't have a default value")
}
