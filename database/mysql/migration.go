/*
Copyright 2022-present The Ztalab Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mysql

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/ztalab/ZACA/pkg/logger"
	_ "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	lo := logger.Named("migration")
	sql, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get DB instance: %v", err)
	}
	driver, err := mysql.WithInstance(sql, &mysql.Config{})
	m, err := migrate.NewWithDatabaseInstance(
		"file://database/mysql/migrations/",
		"mysql", driver)
	if err != nil {
		return fmt.Errorf("migrate instance error: %v", err)
	}
	if err = m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			lo.Info("no changes.")
			return nil
		}
		return fmt.Errorf("MySQL migration exception: %v", err)
	}
	lo.Info("Migrations success.")
	return nil
}
