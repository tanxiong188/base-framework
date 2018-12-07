// Copyright 2018 cloudy 272685110@qq.com.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.
package repositories

import (
	"github.com/hashicorp/go-version"
	"github.com/itcloudy/base-framework/pkg/models"
)

type IMigrationHistoryRepository interface {
	//获得已升级的最新版本
	CurrentVersion() (string, error)
	//升级到某个版本，若中间存在多个，则中间版本同样升级
	ApplyMigrations(collection version.Collection, migrates map[string]string) (err error)
	//列出所有的版本，包括系统中存在的没有安装的
	ListMigration() (migrates []models.MigrationHistory, err error)
}