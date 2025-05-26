package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/gorm"
)

// Migrator 数据库迁移管理器
type Migrator struct {
	db *gorm.DB
}

// NewMigrator 创建迁移管理器
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Migrate 执行数据库迁移
func (m *Migrator) Migrate(migrationsPath string) error {
	// 1. 创建迁移记录表
	if err := m.createMigrationTable(); err != nil {
		return fmt.Errorf("创建迁移记录表失败: %v", err)
	}

	// 2. 获取已执行的迁移
	executedMigrations, err := m.getExecutedMigrations()
	if err != nil {
		return fmt.Errorf("获取已执行迁移失败: %v", err)
	}

	// 3. 读取迁移文件
	files, err := os.ReadDir(migrationsPath)
	if err != nil {
		return fmt.Errorf("读取迁移文件目录失败: %v", err)
	}

	// 4. 过滤并排序迁移文件
	var migrationFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".up.sql") {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}
	sort.Strings(migrationFiles)

	// 5. 执行未执行的迁移
	for _, file := range migrationFiles {
		migrationName := strings.TrimSuffix(file, ".up.sql")
		if _, exists := executedMigrations[migrationName]; exists {
			continue
		}

		// 读取迁移文件内容
		content, err := os.ReadFile(filepath.Join(migrationsPath, file))
		if err != nil {
			return fmt.Errorf("读取迁移文件 %s 失败: %v", file, err)
		}

		// 执行迁移
		if err := m.db.Exec(string(content)).Error; err != nil {
			return fmt.Errorf("执行迁移 %s 失败: %v", migrationName, err)
		}

		// 记录迁移
		if err := m.recordMigration(migrationName); err != nil {
			return fmt.Errorf("记录迁移 %s 失败: %v", migrationName, err)
		}

		fmt.Printf("执行迁移: %s\n", migrationName)
	}

	return nil
}

// Rollback 回滚最后一次迁移
func (m *Migrator) Rollback(migrationsPath string) error {
	// 1. 获取最后一次执行的迁移
	var lastMigration string
	if err := m.db.Table("migrations").
		Order("id DESC").
		Limit(1).
		Pluck("migration", &lastMigration).Error; err != nil {
		return fmt.Errorf("获取最后一次迁移失败: %v", err)
	}

	if lastMigration == "" {
		return fmt.Errorf("没有可回滚的迁移")
	}

	// 2. 读取回滚文件
	rollbackFile := filepath.Join(migrationsPath, lastMigration+".down.sql")
	content, err := os.ReadFile(rollbackFile)
	if err != nil {
		return fmt.Errorf("读取回滚文件失败: %v", err)
	}

	// 3. 执行回滚
	if err := m.db.Exec(string(content)).Error; err != nil {
		return fmt.Errorf("执行回滚失败: %v", err)
	}

	// 4. 删除迁移记录
	if err := m.db.Table("migrations").
		Where("migration = ?", lastMigration).
		Delete(&struct{}{}).Error; err != nil {
		return fmt.Errorf("删除迁移记录失败: %v", err)
	}

	fmt.Printf("回滚迁移: %s\n", lastMigration)
	return nil
}

// createMigrationTable 创建迁移记录表
func (m *Migrator) createMigrationTable() error {
	return m.db.Exec(`
		CREATE TABLE IF NOT EXISTS migrations (
			id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
			migration VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id),
			UNIQUE KEY uk_migration (migration)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
	`).Error
}

// getExecutedMigrations 获取已执行的迁移
func (m *Migrator) getExecutedMigrations() (map[string]struct{}, error) {
	var migrations []string
	if err := m.db.Table("migrations").Pluck("migration", &migrations).Error; err != nil {
		return nil, err
	}

	executed := make(map[string]struct{})
	for _, migration := range migrations {
		executed[migration] = struct{}{}
	}
	return executed, nil
}

// recordMigration 记录迁移
func (m *Migrator) recordMigration(migration string) error {
	return m.db.Exec("INSERT INTO migrations (migration) VALUES (?)", migration).Error
}
