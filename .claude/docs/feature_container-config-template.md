# Container Configuration Template System

## 목적 및 배경

### 문제 정의
현재 Burndler는 개발자가 직접 Container 설정 파일들을 수동으로 편집하여 인스톨러를 제작합니다. 이 과정은:
- 현장 엔지니어에게 교육시키기 어려움
- MSA 구조에서 여러 Container의 연관 설정을 관리하기 복잡함
- 설정 실수로 인한 배포 실패 위험이 높음
- 고객별 맞춤 설정에 많은 시간이 소요됨

### 해결 방안
**지능형 템플릿 시스템**을 도입하여:
1. **개발자**: 설정 템플릿을 생성하고 UI 스키마를 정의
2. **현장 엔지니어**: GUI를 통해 기능 단위로 설정을 쉽게 변경
3. **시스템**: 설정 간 의존성을 자동 검증하고 파일 구조를 실시간 시각화
4. **빌드**: 템플릿을 렌더링하고 조건부 파일/에셋을 자동으로 포함/제외

### 핵심 기능
- 다양한 포맷(YAML, JSON, ENV) 템플릿 지원
- 설정 간 의존성 자동 검증
- 실시간 파일 구조 시각화
- 조건부 파일/에셋 포함/제외
- 대용량 에셋의 임베드/다운로드 선택

---

## Phase 1: 데이터베이스 스키마 및 백엔드 기반

**목표**: 템플릿 시스템의 데이터 구조와 핵심 백엔드 서비스를 구축합니다.

**예상 기간**: 2-3주

### Task 1.1: 데이터베이스 스키마 생성 (완료)

**목적**: 템플릿, 파일, 에셋을 저장할 데이터베이스 테이블을 추가합니다.

**구현 내용**:
```sql
-- 1. container_configurations 테이블
-- ContainerVersion별 설정 템플릿 메타데이터
CREATE TABLE container_configurations (
    id SERIAL PRIMARY KEY,
    container_version_id INTEGER NOT NULL REFERENCES container_versions(id) ON DELETE CASCADE,
    ui_schema JSONB,                    -- 프론트엔드 설정 UI 스키마
    dependency_rules JSONB,             -- 의존성 규칙들
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(container_version_id)
);

-- 2. container_files 테이블
-- 템플릿 파일 및 정적 파일 관리
CREATE TABLE container_files (
    id SERIAL PRIMARY KEY,
    container_version_id INTEGER NOT NULL REFERENCES container_versions(id) ON DELETE CASCADE,
    file_path VARCHAR(512) NOT NULL,   -- 인스톨러 내 경로 (예: "config/app.yaml")
    file_type VARCHAR(20) NOT NULL,    -- 'template', 'static'
    storage_path VARCHAR(512),          -- Storage 실제 경로
    template_format VARCHAR(20),        -- 'yaml', 'json', 'env', 'text'
    display_condition TEXT,             -- 표시 조건 (템플릿 표현식)
    is_directory BOOLEAN DEFAULT FALSE,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 3. container_assets 테이블
-- 에셋 파일 메타데이터 및 다운로드 정보
CREATE TABLE container_assets (
    id SERIAL PRIMARY KEY,
    container_version_id INTEGER NOT NULL REFERENCES container_versions(id) ON DELETE CASCADE,
    original_file_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(512) NOT NULL,   -- 인스톨러 내 경로
    asset_type VARCHAR(20) NOT NULL,   -- 'config', 'data', 'script', 'binary', 'document'
    mime_type VARCHAR(100),
    file_size BIGINT NOT NULL,
    checksum VARCHAR(64) NOT NULL,     -- SHA256
    compressed BOOLEAN DEFAULT FALSE,
    include_condition TEXT,             -- 포함 조건 (템플릿 표현식)
    storage_type VARCHAR(20) NOT NULL, -- 'embedded', 'download'
    storage_path VARCHAR(512) NOT NULL,
    download_url TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 4. service_configurations 테이블
-- Service별 Container 설정 값 저장
CREATE TABLE service_configurations (
    id SERIAL PRIMARY KEY,
    service_id INTEGER NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    container_id INTEGER NOT NULL REFERENCES containers(id) ON DELETE CASCADE,
    configuration_values JSONB,         -- 사용자가 설정한 값들
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(service_id, container_id)
);

-- 인덱스 생성
CREATE INDEX idx_container_files_version ON container_files(container_version_id);
CREATE INDEX idx_container_assets_version ON container_assets(container_version_id);
CREATE INDEX idx_service_configurations_service ON service_configurations(service_id);
CREATE INDEX idx_service_configurations_container ON service_configurations(container_id);
```

**파일 경로**:
- `backend/internal/app/migration.go` - 마이그레이션 함수 추가
- 새 마이그레이션 파일 생성 (예: `migrations/007_add_configuration_tables.sql`)

**검증 기준**:
- [ ] 마이그레이션이 오류 없이 실행됨
- [ ] 모든 테이블과 인덱스가 정상 생성됨
- [ ] Foreign Key 제약조건이 올바르게 설정됨
- [ ] `make test-unit` 통과

**테스트**:
```bash
# 마이그레이션 테스트
make dev-backend
# 데이터베이스 확인
psql -d burndler -c "\dt container_*"
psql -d burndler -c "\dt service_configurations"
```

---

### Task 1.2: Go 모델 정의 (완료)

**목적**: 새로운 테이블에 대응하는 Go 구조체를 정의합니다.

**구현 내용**:
```go
// backend/internal/models/container_configuration.go
package models

import (
    "time"
    "gorm.io/datatypes"
    "gorm.io/gorm"
)

type ContainerConfiguration struct {
    ID                 uint           `gorm:"primaryKey" json:"id"`
    ContainerVersionID uint           `gorm:"not null;uniqueIndex" json:"container_version_id"`
    UISchema           datatypes.JSON `gorm:"type:jsonb" json:"ui_schema"`
    DependencyRules    datatypes.JSON `gorm:"type:jsonb" json:"dependency_rules"`
    CreatedAt          time.Time      `json:"created_at"`
    UpdatedAt          time.Time      `json:"updated_at"`

    // Relationships
    ContainerVersion ContainerVersion `gorm:"foreignKey:ContainerVersionID" json:"container_version,omitempty"`
    Files            []ContainerFile  `gorm:"foreignKey:ContainerVersionID;references:ContainerVersionID" json:"files,omitempty"`
    Assets           []ContainerAsset `gorm:"foreignKey:ContainerVersionID;references:ContainerVersionID" json:"assets,omitempty"`
}

func (ContainerConfiguration) TableName() string {
    return "container_configurations"
}

type ContainerFile struct {
    ID                 uint      `gorm:"primaryKey" json:"id"`
    ContainerVersionID uint      `gorm:"not null;index" json:"container_version_id"`
    FilePath           string    `gorm:"size:512;not null" json:"file_path"`
    FileType           string    `gorm:"size:20;not null" json:"file_type"` // 'template', 'static'
    StoragePath        string    `gorm:"size:512" json:"storage_path"`
    TemplateFormat     string    `gorm:"size:20" json:"template_format"` // 'yaml', 'json', 'env', 'text'
    DisplayCondition   string    `gorm:"type:text" json:"display_condition"`
    IsDirectory        bool      `gorm:"default:false" json:"is_directory"`
    Description        string    `gorm:"type:text" json:"description"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}

func (ContainerFile) TableName() string {
    return "container_files"
}

type ContainerAsset struct {
    ID                 uint      `gorm:"primaryKey" json:"id"`
    ContainerVersionID uint      `gorm:"not null;index" json:"container_version_id"`
    OriginalFileName   string    `gorm:"size:255;not null" json:"original_file_name"`
    FilePath           string    `gorm:"size:512;not null" json:"file_path"`
    AssetType          string    `gorm:"size:20;not null" json:"asset_type"` // 'config', 'data', 'script', 'binary', 'document'
    MimeType           string    `gorm:"size:100" json:"mime_type"`
    FileSize           int64     `gorm:"not null" json:"file_size"`
    Checksum           string    `gorm:"size:64;not null" json:"checksum"` // SHA256
    Compressed         bool      `gorm:"default:false" json:"compressed"`
    IncludeCondition   string    `gorm:"type:text" json:"include_condition"`
    StorageType        string    `gorm:"size:20;not null" json:"storage_type"` // 'embedded', 'download'
    StoragePath        string    `gorm:"size:512;not null" json:"storage_path"`
    DownloadURL        string    `gorm:"type:text" json:"download_url"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`
}

func (ContainerAsset) TableName() string {
    return "container_assets"
}

type ServiceConfiguration struct {
    ID                  uint           `gorm:"primaryKey" json:"id"`
    ServiceID           uint           `gorm:"not null;uniqueIndex:idx_service_container" json:"service_id"`
    ContainerID         uint           `gorm:"not null;uniqueIndex:idx_service_container" json:"container_id"`
    ConfigurationValues datatypes.JSON `gorm:"type:jsonb" json:"configuration_values"`
    CreatedAt           time.Time      `json:"created_at"`
    UpdatedAt           time.Time      `json:"updated_at"`

    // Relationships
    Service   Service   `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
    Container Container `gorm:"foreignKey:ContainerID" json:"container,omitempty"`
}

func (ServiceConfiguration) TableName() string {
    return "service_configurations"
}
```

**파일 경로**:
- `backend/internal/models/container_configuration.go` (신규)

**검증 기준**:
- [ ] GORM 태그가 올바르게 설정됨
- [ ] JSON 직렬화/역직렬화가 정상 동작
- [ ] 모델 간 관계가 올바르게 정의됨
- [ ] `make lint-backend` 통과

**테스트**:
```go
// backend/internal/models/container_configuration_test.go
func TestContainerConfiguration(t *testing.T) {
    db := setupTestDB(t)

    // ContainerConfiguration 생성 테스트
    config := &ContainerConfiguration{
        ContainerVersionID: 1,
        UISchema:           datatypes.JSON([]byte(`{"fields":[]}`)),
        DependencyRules:    datatypes.JSON([]byte(`{"rules":[]}`)),
    }

    err := db.Create(config).Error
    assert.NoError(t, err)
    assert.NotZero(t, config.ID)
}
```

---

### Task 1.3: 템플릿 엔진 구현 (완료)

**목적**: 다양한 포맷(YAML, JSON, ENV)의 템플릿을 렌더링하는 엔진을 구현합니다.

**구현 내용**:
```go
// backend/internal/services/template_engine.go
package services

import (
    "bytes"
    "encoding/json"
    "fmt"
    "strings"
    "text/template"

    "gopkg.in/yaml.v3"
)

type TemplateEngine struct {
    funcMap template.FuncMap
}

func NewTemplateEngine() *TemplateEngine {
    return &TemplateEngine{
        funcMap: getTemplateFuncMap(),
    }
}

// RenderYAML renders YAML template with structure preservation
func (te *TemplateEngine) RenderYAML(templateContent string, variables map[string]interface{}) (string, error) {
    // 1. Go template 렌더링
    tmpl, err := template.New("yaml").Funcs(te.funcMap).Parse(templateContent)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, variables); err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }

    // 2. YAML 구조 검증
    var yamlData interface{}
    if err := yaml.Unmarshal(buf.Bytes(), &yamlData); err != nil {
        return "", fmt.Errorf("invalid YAML after rendering: %w", err)
    }

    // 3. 포맷팅된 YAML 반환
    formatted, err := yaml.Marshal(yamlData)
    if err != nil {
        return "", err
    }

    return string(formatted), nil
}

// RenderJSON renders JSON template with structure preservation
func (te *TemplateEngine) RenderJSON(templateContent string, variables map[string]interface{}) (string, error) {
    // 1. Go template 렌더링
    tmpl, err := template.New("json").Funcs(te.funcMap).Parse(templateContent)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, variables); err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }

    // 2. JSON 구조 검증
    var jsonData interface{}
    if err := json.Unmarshal(buf.Bytes(), &jsonData); err != nil {
        return "", fmt.Errorf("invalid JSON after rendering: %w", err)
    }

    // 3. 포맷팅된 JSON 반환
    formatted, err := json.MarshalIndent(jsonData, "", "  ")
    if err != nil {
        return "", err
    }

    return string(formatted), nil
}

// RenderEnv renders ENV file template
func (te *TemplateEngine) RenderEnv(templateContent string, variables map[string]interface{}) (string, error) {
    tmpl, err := template.New("env").Funcs(te.funcMap).Parse(templateContent)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, variables); err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }

    return buf.String(), nil
}

// RenderText renders plain text template
func (te *TemplateEngine) RenderText(templateContent string, variables map[string]interface{}) (string, error) {
    tmpl, err := template.New("text").Funcs(te.funcMap).Parse(templateContent)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, variables); err != nil {
        return "", fmt.Errorf("template execution error: %w", err)
    }

    return buf.String(), nil
}

// Render automatically selects renderer based on format
func (te *TemplateEngine) Render(templateContent string, format string, variables map[string]interface{}) (string, error) {
    switch format {
    case "yaml":
        return te.RenderYAML(templateContent, variables)
    case "json":
        return te.RenderJSON(templateContent, variables)
    case "env":
        return te.RenderEnv(templateContent, variables)
    case "text":
        return te.RenderText(templateContent, variables)
    default:
        return "", fmt.Errorf("unsupported template format: %s", format)
    }
}

// getTemplateFuncMap returns template helper functions
func getTemplateFuncMap() template.FuncMap {
    return template.FuncMap{
        // String functions
        "upper":     strings.ToUpper,
        "lower":     strings.ToLower,
        "trim":      strings.TrimSpace,
        "replace":   strings.ReplaceAll,

        // Default value
        "default": func(defaultVal, val interface{}) interface{} {
            if val == nil || val == "" {
                return defaultVal
            }
            return val
        },

        // Math functions
        "add": func(a, b int) int { return a + b },
        "sub": func(a, b int) int { return a - b },

        // Conditionals
        "eq": func(a, b interface{}) bool { return a == b },
        "ne": func(a, b interface{}) bool { return a != b },
    }
}
```

**파일 경로**:
- `backend/internal/services/template_engine.go` (신규)
- `backend/internal/services/template_engine_test.go` (신규)

**검증 기준**:
- [ ] YAML 템플릿이 올바르게 렌더링됨
- [ ] JSON 템플릿이 올바르게 렌더링됨
- [ ] ENV 템플릿이 올바르게 렌더링됨
- [ ] 잘못된 템플릿 문법에 대한 오류 처리
- [ ] 모든 단위 테스트 통과

**테스트**:
```go
// backend/internal/services/template_engine_test.go
func TestTemplateEngine_RenderYAML(t *testing.T) {
    engine := NewTemplateEngine()

    template := `
database:
  host: {{ .Database.Host }}
  port: {{ .Database.Port | default 5432 }}
  name: {{ .Database.Name }}
`

    variables := map[string]interface{}{
        "Database": map[string]interface{}{
            "Host": "localhost",
            "Name": "testdb",
        },
    }

    result, err := engine.RenderYAML(template, variables)
    assert.NoError(t, err)
    assert.Contains(t, result, "host: localhost")
    assert.Contains(t, result, "port: 5432")
    assert.Contains(t, result, "name: testdb")
}
```

---

### Task 1.4: 기본 API 엔드포인트 구현 (완료)

**목적**: Container Configuration 생성/조회/수정/삭제 API를 구현합니다.

**구현 내용**:
```go
// backend/internal/handlers/container_configuration.go
package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "burndler/internal/models"
    "gorm.io/gorm"
)

type ContainerConfigurationHandler struct {
    db *gorm.DB
}

func NewContainerConfigurationHandler(db *gorm.DB) *ContainerConfigurationHandler {
    return &ContainerConfigurationHandler{db: db}
}

// CreateConfiguration creates a new container configuration
func (h *ContainerConfigurationHandler) CreateConfiguration(c *gin.Context) {
    containerID, _ := strconv.ParseUint(c.Param("container_id"), 10, 64)
    versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 64)

    var req struct {
        UISchema        json.RawMessage `json:"ui_schema"`
        DependencyRules json.RawMessage `json:"dependency_rules"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Verify container version exists
    var version models.ContainerVersion
    if err := h.db.Where("id = ? AND container_id = ?", versionID, containerID).First(&version).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Container version not found"})
        return
    }

    config := &models.ContainerConfiguration{
        ContainerVersionID: uint(versionID),
        UISchema:           req.UISchema,
        DependencyRules:    req.DependencyRules,
    }

    if err := h.db.Create(config).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create configuration"})
        return
    }

    c.JSON(http.StatusCreated, config)
}

// GetConfiguration retrieves a container configuration
func (h *ContainerConfigurationHandler) GetConfiguration(c *gin.Context) {
    versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 64)

    var config models.ContainerConfiguration
    if err := h.db.Where("container_version_id = ?", versionID).
        Preload("Files").
        Preload("Assets").
        First(&config).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve configuration"})
        return
    }

    c.JSON(http.StatusOK, config)
}

// UpdateConfiguration updates a container configuration
func (h *ContainerConfigurationHandler) UpdateConfiguration(c *gin.Context) {
    versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 64)

    var req struct {
        UISchema        json.RawMessage `json:"ui_schema"`
        DependencyRules json.RawMessage `json:"dependency_rules"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    var config models.ContainerConfiguration
    if err := h.db.Where("container_version_id = ?", versionID).First(&config).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Configuration not found"})
        return
    }

    config.UISchema = req.UISchema
    config.DependencyRules = req.DependencyRules

    if err := h.db.Save(&config).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update configuration"})
        return
    }

    c.JSON(http.StatusOK, config)
}

// DeleteConfiguration deletes a container configuration
func (h *ContainerConfigurationHandler) DeleteConfiguration(c *gin.Context) {
    versionID, _ := strconv.ParseUint(c.Param("version_id"), 10, 64)

    if err := h.db.Where("container_version_id = ?", versionID).Delete(&models.ContainerConfiguration{}).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete configuration"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Configuration deleted successfully"})
}
```

**라우트 등록**:
```go
// backend/internal/server/server.go에 추가
configHandler := handlers.NewContainerConfigurationHandler(s.DB)

// Developer only routes
api.POST("/containers/:container_id/versions/:version_id/configuration",
    middleware.RequireRole("Developer"),
    configHandler.CreateConfiguration)
api.PUT("/containers/:container_id/versions/:version_id/configuration",
    middleware.RequireRole("Developer"),
    configHandler.UpdateConfiguration)
api.DELETE("/containers/:container_id/versions/:version_id/configuration",
    middleware.RequireRole("Developer"),
    configHandler.DeleteConfiguration)

// Read access for all authenticated users
api.GET("/containers/:container_id/versions/:version_id/configuration",
    configHandler.GetConfiguration)
```

**파일 경로**:
- `backend/internal/handlers/container_configuration.go` (신규)
- `backend/internal/handlers/container_configuration_test.go` (신규)
- `backend/internal/server/server.go` (수정)

**검증 기준**:
- [ ] POST /api/v1/containers/{id}/versions/{version}/configuration 동작
- [ ] GET /api/v1/containers/{id}/versions/{version}/configuration 동작
- [ ] PUT /api/v1/containers/{id}/versions/{version}/configuration 동작
- [ ] DELETE /api/v1/containers/{id}/versions/{version}/configuration 동작
- [ ] RBAC 권한이 올바르게 적용됨
- [ ] 모든 API 테스트 통과

**테스트**:
```bash
# API 테스트
curl -X POST http://localhost:8080/api/v1/containers/1/versions/1/configuration \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"ui_schema": {"fields": []}, "dependency_rules": {"rules": []}}'
```

---

## Phase 2: 프론트엔드 UI 및 실시간 미리보기

**목표**: 현장 엔지니어가 GUI로 설정할 수 있는 UI와 파일 구조 시각화를 구현합니다.

**예상 기간**: 2-3주

### Task 2.1: 설정 UI 스키마 구조 정의 (완료)

**목적**: 백엔드에서 정의한 UI 스키마를 프론트엔드에서 렌더링할 수 있도록 TypeScript 타입을 정의합니다.

**구현 내용**:
```typescript
// frontend/src/types/configuration.ts
export interface UISchema {
  sections: UISection[]
}

export interface UISection {
  id: string
  title: string
  description?: string
  fields: UIField[]
  condition?: string  // 섹션 표시 조건
}

export interface UIField {
  key: string                    // 변수 키 (예: "Database.Host")
  type: UIFieldType
  label: string
  description?: string
  defaultValue?: any
  required?: boolean
  validation?: FieldValidation
  affects?: string[]             // 영향받는 파일 경로들
  dependencies?: string[]        // 의존하는 다른 필드들
  ui?: FieldUIOptions
}

export type UIFieldType =
  | 'boolean'
  | 'string'
  | 'number'
  | 'select'
  | 'multiselect'
  | 'textarea'

export interface FieldValidation {
  min?: number
  max?: number
  minLength?: number
  maxLength?: number
  pattern?: string
  enum?: string[]
  custom?: string  // 커스텀 검증 표현식
}

export interface FieldUIOptions {
  placeholder?: string
  helpText?: string
  options?: SelectOption[]  // select/multiselect용
  rows?: number             // textarea용
  unit?: string             // 단위 표시 (예: "MB", "초")
}

export interface SelectOption {
  label: string
  value: string | number
}

export interface ConfigurationValues {
  [key: string]: any
}

export interface DependencyRule {
  type: 'requires' | 'conflicts' | 'cascades'
  field: string
  condition: string
  target: string
  targetValue?: any
  message?: string
}
```

**파일 경로**:
- `frontend/src/types/configuration.ts` (신규)

**검증 기준**:
- [ ] TypeScript 컴파일 오류 없음
- [ ] 백엔드 스키마와 일치함
- [ ] `npm run lint` 통과

---

### Task 2.2: 설정 폼 컴포넌트 구현 (완료)

**목적**: UI 스키마를 기반으로 동적으로 폼을 렌더링하는 컴포넌트를 구현합니다.

**구현 내용**:
```typescript
// frontend/src/components/configuration/ConfigurationForm.tsx
import React, { useState, useCallback } from 'react'
import { UISchema, ConfigurationValues } from '@/types/configuration'
import { ConfigurationSection } from './ConfigurationSection'

interface ConfigurationFormProps {
  schema: UISchema
  initialValues?: ConfigurationValues
  onChange: (values: ConfigurationValues) => void
  onValidate?: (errors: ValidationErrors) => void
}

export const ConfigurationForm: React.FC<ConfigurationFormProps> = ({
  schema,
  initialValues = {},
  onChange,
  onValidate
}) => {
  const [values, setValues] = useState<ConfigurationValues>(initialValues)
  const [errors, setErrors] = useState<ValidationErrors>({})

  const handleFieldChange = useCallback((key: string, value: any) => {
    const newValues = {
      ...values,
      [key]: value
    }

    setValues(newValues)
    onChange(newValues)

    // Validate field
    const fieldErrors = validateField(key, value, schema)
    setErrors(prev => ({
      ...prev,
      [key]: fieldErrors
    }))

    if (onValidate) {
      onValidate(errors)
    }
  }, [values, schema, onChange, onValidate, errors])

  const evaluateCondition = useCallback((condition: string | undefined): boolean => {
    if (!condition) return true

    try {
      // Simple condition evaluation
      // 예: "Database.Enabled === true"
      return new Function('values', `with(values) { return ${condition} }`)(values)
    } catch {
      return false
    }
  }, [values])

  return (
    <div className="configuration-form space-y-6">
      {schema.sections.map(section => {
        const shouldShow = evaluateCondition(section.condition)

        if (!shouldShow) return null

        return (
          <ConfigurationSection
            key={section.id}
            section={section}
            values={values}
            errors={errors}
            onChange={handleFieldChange}
            evaluateCondition={evaluateCondition}
          />
        )
      })}
    </div>
  )
}
```

```typescript
// frontend/src/components/configuration/ConfigurationSection.tsx
import React from 'react'
import { UISection, ConfigurationValues, ValidationErrors } from '@/types/configuration'
import { ConfigurationField } from './ConfigurationField'

interface ConfigurationSectionProps {
  section: UISection
  values: ConfigurationValues
  errors: ValidationErrors
  onChange: (key: string, value: any) => void
  evaluateCondition: (condition: string | undefined) => boolean
}

export const ConfigurationSection: React.FC<ConfigurationSectionProps> = ({
  section,
  values,
  errors,
  onChange,
  evaluateCondition
}) => {
  return (
    <div className="section border rounded-lg p-6 bg-white">
      <h3 className="text-lg font-semibold mb-2">{section.title}</h3>
      {section.description && (
        <p className="text-sm text-gray-600 mb-4">{section.description}</p>
      )}

      <div className="fields space-y-4">
        {section.fields.map(field => {
          const shouldShow = evaluateCondition(field.condition)
          const isDisabled = !evaluateCondition(field.enabledCondition)

          if (!shouldShow) return null

          return (
            <ConfigurationField
              key={field.key}
              field={field}
              value={values[field.key]}
              error={errors[field.key]}
              disabled={isDisabled}
              onChange={(value) => onChange(field.key, value)}
            />
          )
        })}
      </div>
    </div>
  )
}
```

```typescript
// frontend/src/components/configuration/ConfigurationField.tsx
import React from 'react'
import { UIField } from '@/types/configuration'

interface ConfigurationFieldProps {
  field: UIField
  value: any
  error?: string
  disabled?: boolean
  onChange: (value: any) => void
}

export const ConfigurationField: React.FC<ConfigurationFieldProps> = ({
  field,
  value,
  error,
  disabled,
  onChange
}) => {
  const renderField = () => {
    switch (field.type) {
      case 'boolean':
        return (
          <label className="flex items-center space-x-2">
            <input
              type="checkbox"
              checked={value || false}
              disabled={disabled}
              onChange={(e) => onChange(e.target.checked)}
              className="form-checkbox"
            />
            <span>{field.label}</span>
          </label>
        )

      case 'string':
        return (
          <div>
            <label className="block text-sm font-medium mb-1">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <input
              type="text"
              value={value || ''}
              disabled={disabled}
              placeholder={field.ui?.placeholder}
              onChange={(e) => onChange(e.target.value)}
              className="form-input w-full"
            />
          </div>
        )

      case 'number':
        return (
          <div>
            <label className="block text-sm font-medium mb-1">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <div className="flex items-center">
              <input
                type="number"
                value={value || ''}
                disabled={disabled}
                min={field.validation?.min}
                max={field.validation?.max}
                onChange={(e) => onChange(Number(e.target.value))}
                className="form-input w-full"
              />
              {field.ui?.unit && (
                <span className="ml-2 text-sm text-gray-500">{field.ui.unit}</span>
              )}
            </div>
          </div>
        )

      case 'select':
        return (
          <div>
            <label className="block text-sm font-medium mb-1">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <select
              value={value || ''}
              disabled={disabled}
              onChange={(e) => onChange(e.target.value)}
              className="form-select w-full"
            >
              <option value="">선택하세요</option>
              {field.ui?.options?.map(opt => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </div>
        )

      case 'textarea':
        return (
          <div>
            <label className="block text-sm font-medium mb-1">
              {field.label}
              {field.required && <span className="text-red-500 ml-1">*</span>}
            </label>
            <textarea
              value={value || ''}
              disabled={disabled}
              rows={field.ui?.rows || 3}
              placeholder={field.ui?.placeholder}
              onChange={(e) => onChange(e.target.value)}
              className="form-textarea w-full"
            />
          </div>
        )

      default:
        return null
    }
  }

  return (
    <div className={`field ${disabled ? 'opacity-50' : ''}`}>
      {renderField()}
      {field.description && !error && (
        <p className="text-xs text-gray-500 mt-1">{field.description}</p>
      )}
      {error && (
        <p className="text-xs text-red-500 mt-1">{error}</p>
      )}
      {field.ui?.helpText && (
        <p className="text-xs text-blue-500 mt-1">{field.ui.helpText}</p>
      )}
    </div>
  )
}
```

**파일 경로**:
- `frontend/src/components/configuration/ConfigurationForm.tsx` (신규)
- `frontend/src/components/configuration/ConfigurationSection.tsx` (신규)
- `frontend/src/components/configuration/ConfigurationField.tsx` (신규)

**검증 기준**:
- [ ] UI 스키마를 기반으로 폼이 동적 생성됨
- [ ] 모든 필드 타입이 올바르게 렌더링됨
- [ ] 필드 값 변경이 정상 동작함
- [ ] 조건부 필드 표시/숨김이 동작함

---

### Task 2.3: 파일 구조 시각화 컴포넌트 (완료)

**목적**: 설정에 따라 포함될 파일들을 트리 구조로 시각화합니다.

**구현 내용**:
```typescript
// frontend/src/types/fileStructure.ts
export interface FileStructureNode {
  name: string
  path: string
  type: 'file' | 'directory'
  fileType?: 'template' | 'asset' | 'static'
  condition?: string
  visible: boolean
  description?: string
  size?: number
  children?: FileStructureNode[]
  isGenerated?: boolean
}

export interface FileStructureState {
  rootNodes: FileStructureNode[]
  totalFiles: number
  totalSize: number
  visibleFiles: number
  hiddenFiles: number
}
```

```typescript
// frontend/src/components/configuration/FileStructureViewer.tsx
import React, { useMemo } from 'react'
import { FileStructureState, FileStructureNode } from '@/types/fileStructure'
import { FileTreeNode } from './FileTreeNode'
import { formatFileSize } from '@/utils/format'

interface FileStructureViewerProps {
  structure: FileStructureState
}

export const FileStructureViewer: React.FC<FileStructureViewerProps> = ({
  structure
}) => {
  return (
    <div className="file-structure-viewer border rounded-lg p-4 bg-gray-50">
      <div className="structure-header mb-4">
        <h3 className="text-lg font-semibold mb-2">📁 파일 구조 미리보기</h3>
        <div className="structure-stats flex space-x-4 text-sm">
          <span className="flex items-center">
            <span className="font-medium mr-1">📄</span>
            {structure.visibleFiles}개 파일
          </span>
          <span className="flex items-center">
            <span className="font-medium mr-1">📦</span>
            {formatFileSize(structure.totalSize)}
          </span>
          {structure.hiddenFiles > 0 && (
            <span className="text-gray-500">
              ({structure.hiddenFiles}개 숨김)
            </span>
          )}
        </div>
      </div>

      <div className="structure-tree bg-white border rounded p-3 max-h-96 overflow-y-auto">
        {structure.rootNodes.map(node => (
          <FileTreeNode key={node.path} node={node} level={0} />
        ))}
      </div>
    </div>
  )
}
```

```typescript
// frontend/src/components/configuration/FileTreeNode.tsx
import React, { useState } from 'react'
import { FileStructureNode } from '@/types/fileStructure'
import { formatFileSize } from '@/utils/format'

interface FileTreeNodeProps {
  node: FileStructureNode
  level: number
}

export const FileTreeNode: React.FC<FileTreeNodeProps> = ({ node, level }) => {
  const [expanded, setExpanded] = useState(true)

  const getNodeIcon = () => {
    if (node.type === 'directory') {
      return expanded ? '📁' : '📂'
    }

    switch (node.fileType) {
      case 'template': return '📝'
      case 'asset': return '🗂️'
      case 'static': return '📄'
      default: return '📄'
    }
  }

  const getStatusClass = () => {
    if (!node.condition) return 'text-gray-900'
    return node.visible ? 'text-green-700' : 'text-gray-400 line-through'
  }

  return (
    <div className="file-node" style={{ marginLeft: `${level * 20}px` }}>
      <div
        className={`node-content flex items-center py-1 px-2 hover:bg-gray-50 rounded cursor-pointer ${getStatusClass()}`}
        onClick={() => node.type === 'directory' && setExpanded(!expanded)}
      >
        <span className="node-icon mr-2">{getNodeIcon()}</span>
        <span className="node-name flex-1 text-sm">{node.name}</span>

        {node.size && (
          <span className="node-size text-xs text-gray-500 mr-2">
            {formatFileSize(node.size)}
          </span>
        )}

        {node.condition && (
          <span className={`condition-badge text-xs px-2 py-1 rounded ${
            node.visible ? 'bg-green-100 text-green-700' : 'bg-gray-100 text-gray-500'
          }`}>
            {node.visible ? '✅' : '❌'} 조건부
          </span>
        )}

        {node.isGenerated && (
          <span className="generated-badge text-xs px-2 py-1 rounded bg-blue-100 text-blue-700 ml-2">
            🔄 생성됨
          </span>
        )}
      </div>

      {node.description && (
        <div className="node-description text-xs text-gray-600 ml-8 mb-1">
          {node.description}
        </div>
      )}

      {node.condition && (
        <div className="node-condition text-xs text-gray-500 ml-8 mb-1">
          조건: <code className="bg-gray-100 px-1 rounded">{node.condition}</code>
        </div>
      )}

      {node.type === 'directory' && expanded && node.children && (
        <div className="node-children">
          {node.children.map(child => (
            <FileTreeNode key={child.path} node={child} level={level + 1} />
          ))}
        </div>
      )}
    </div>
  )
}
```

**파일 경로**:
- `frontend/src/types/fileStructure.ts` (신규)
- `frontend/src/components/configuration/FileStructureViewer.tsx` (신규)
- `frontend/src/components/configuration/FileTreeNode.tsx` (신규)
- `frontend/src/utils/format.ts` (신규 또는 수정)

**검증 기준**:
- [ ] 파일 트리가 계층 구조로 표시됨
- [ ] 폴더 확장/축소가 동작함
- [ ] 조건부 파일이 올바르게 표시됨
- [ ] 파일 크기와 아이콘이 올바르게 표시됨

---

### Task 2.4: 설정 페이지 통합 (완료)

**목적**: 설정 폼과 파일 구조 뷰어를 하나의 페이지에 통합합니다.

**구현 내용**:
```typescript
// frontend/src/pages/ServiceConfigurationPage.tsx
import React, { useState, useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { ConfigurationForm } from '@/components/configuration/ConfigurationForm'
import { FileStructureViewer } from '@/components/configuration/FileStructureViewer'
import { useFileStructureResolver } from '@/hooks/useFileStructureResolver'
import { api } from '@/services/api'

export const ServiceConfigurationPage: React.FC = () => {
  const { serviceId, containerId } = useParams<{ serviceId: string, containerId: string }>()
  const [schema, setSchema] = useState(null)
  const [values, setValues] = useState({})
  const [loading, setLoading] = useState(true)

  // Load configuration schema
  useEffect(() => {
    const loadConfiguration = async () => {
      try {
        const response = await api.get(`/services/${serviceId}/containers/${containerId}/config`)
        setSchema(response.data.ui_schema)
        setValues(response.data.current_values || {})
      } catch (error) {
        console.error('Failed to load configuration:', error)
      } finally {
        setLoading(false)
      }
    }

    loadConfiguration()
  }, [serviceId, containerId])

  // Resolve file structure based on current values
  const fileStructure = useFileStructureResolver(schema, values)

  const handleSave = async () => {
    try {
      await api.put(`/services/${serviceId}/containers/${containerId}/config`, {
        configuration_values: values
      })
      alert('설정이 저장되었습니다')
    } catch (error) {
      console.error('Failed to save configuration:', error)
      alert('설정 저장에 실패했습니다')
    }
  }

  if (loading) {
    return <div>Loading...</div>
  }

  return (
    <div className="service-configuration-page p-6">
      <div className="page-header mb-6">
        <h1 className="text-2xl font-bold">컨테이너 설정</h1>
        <p className="text-gray-600">서비스에 포함될 컨테이너의 설정을 변경합니다</p>
      </div>

      <div className="grid grid-cols-2 gap-6">
        <div className="config-panel">
          <h2 className="text-xl font-semibold mb-4">설정</h2>
          {schema && (
            <ConfigurationForm
              schema={schema}
              initialValues={values}
              onChange={setValues}
            />
          )}

          <div className="actions mt-6 flex space-x-3">
            <button
              onClick={handleSave}
              className="btn btn-primary px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
            >
              저장
            </button>
            <button
              onClick={() => window.history.back()}
              className="btn btn-secondary px-6 py-2 bg-gray-300 text-gray-700 rounded hover:bg-gray-400"
            >
              취소
            </button>
          </div>
        </div>

        <div className="preview-panel">
          <h2 className="text-xl font-semibold mb-4">파일 구조 미리보기</h2>
          <FileStructureViewer structure={fileStructure} />
        </div>
      </div>
    </div>
  )
}
```

```typescript
// frontend/src/hooks/useFileStructureResolver.ts
import { useMemo } from 'react'
import { UISchema, ConfigurationValues } from '@/types/configuration'
import { FileStructureState } from '@/types/fileStructure'

export const useFileStructureResolver = (
  schema: UISchema | null,
  values: ConfigurationValues
): FileStructureState => {
  return useMemo(() => {
    if (!schema) {
      return {
        rootNodes: [],
        totalFiles: 0,
        totalSize: 0,
        visibleFiles: 0,
        hiddenFiles: 0
      }
    }

    // TODO: Implement actual file structure resolution logic
    // This will be implemented in Task 2.5

    return {
      rootNodes: [],
      totalFiles: 0,
      totalSize: 0,
      visibleFiles: 0,
      hiddenFiles: 0
    }
  }, [schema, values])
}
```

**파일 경로**:
- `frontend/src/pages/ServiceConfigurationPage.tsx` (신규)
- `frontend/src/hooks/useFileStructureResolver.ts` (신규)

**검증 기준**:
- [ ] 페이지가 정상적으로 로드됨
- [ ] 설정 폼과 파일 구조가 나란히 표시됨
- [ ] 설정 변경 시 실시간으로 파일 구조가 업데이트됨
- [ ] 저장 버튼이 정상 동작함

---

## Phase 3: 의존성 엔진 및 검증 시스템

**목표**: 설정 간 의존성을 정의하고 실시간으로 검증하는 시스템을 구축합니다.

**예상 기간**: 2주

### Task 3.1: 의존성 검증 엔진 (백엔드) (완료)

**목적**: 복잡한 의존성 규칙을 정의하고 검증하는 백엔드 엔진을 구현합니다.

**구현 내용**:
```go
// backend/internal/services/dependency_checker.go
package services

import (
    "fmt"
    "strings"
)

type DependencyChecker struct{}

func NewDependencyChecker() *DependencyChecker {
    return &DependencyChecker{}
}

type DependencyRule struct {
    Type      string      `json:"type"`      // "requires", "conflicts", "cascades"
    Field     string      `json:"field"`     // 조건 필드
    Condition string      `json:"condition"` // 조건 표현식
    Target    string      `json:"target"`    // 영향받는 필드
    Message   string      `json:"message"`   // 오류 메시지
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Rule    string `json:"rule"`
}

// ValidateConfiguration validates configuration values against dependency rules
func (dc *DependencyChecker) ValidateConfiguration(
    rules []DependencyRule,
    values map[string]interface{},
) []ValidationError {
    errors := []ValidationError{}

    for _, rule := range rules {
        if err := dc.validateRule(rule, values); err != nil {
            errors = append(errors, *err)
        }
    }

    return errors
}

func (dc *DependencyChecker) validateRule(
    rule DependencyRule,
    values map[string]interface{},
) *ValidationError {
    // 1. Evaluate condition
    conditionMet, err := dc.evaluateCondition(rule.Condition, values)
    if err != nil {
        return &ValidationError{
            Field:   rule.Field,
            Message: fmt.Sprintf("Failed to evaluate condition: %v", err),
            Rule:    rule.Type,
        }
    }

    if !conditionMet {
        return nil // Condition not met, rule doesn't apply
    }

    // 2. Validate based on rule type
    switch rule.Type {
    case "requires":
        return dc.validateRequires(rule, values)
    case "conflicts":
        return dc.validateConflicts(rule, values)
    case "cascades":
        // Cascade rules are handled during value setting, not validation
        return nil
    default:
        return &ValidationError{
            Field:   rule.Field,
            Message: fmt.Sprintf("Unknown rule type: %s", rule.Type),
            Rule:    rule.Type,
        }
    }
}

func (dc *DependencyChecker) validateRequires(
    rule DependencyRule,
    values map[string]interface{},
) *ValidationError {
    targetValue := dc.getNestedValue(values, rule.Target)

    if targetValue == nil || targetValue == "" || targetValue == false {
        message := rule.Message
        if message == "" {
            message = fmt.Sprintf("%s requires %s to be set", rule.Field, rule.Target)
        }

        return &ValidationError{
            Field:   rule.Target,
            Message: message,
            Rule:    "requires",
        }
    }

    return nil
}

func (dc *DependencyChecker) validateConflicts(
    rule DependencyRule,
    values map[string]interface{},
) *ValidationError {
    targetValue := dc.getNestedValue(values, rule.Target)

    if targetValue != nil && targetValue != "" && targetValue != false {
        message := rule.Message
        if message == "" {
            message = fmt.Sprintf("%s conflicts with %s", rule.Field, rule.Target)
        }

        return &ValidationError{
            Field:   rule.Target,
            Message: message,
            Rule:    "conflicts",
        }
    }

    return nil
}

// evaluateCondition evaluates a simple condition expression
func (dc *DependencyChecker) evaluateCondition(
    condition string,
    values map[string]interface{},
) (bool, error) {
    // Simple expression evaluator
    // Supports: ==, !=, >, <, >=, <=, &&, ||

    if condition == "" {
        return true, nil
    }

    // Replace field references with actual values
    expr := condition
    for key, value := range values {
        placeholder := fmt.Sprintf("{{.%s}}", key)
        var replacement string

        switch v := value.(type) {
        case bool:
            replacement = fmt.Sprintf("%t", v)
        case string:
            replacement = fmt.Sprintf(`"%s"`, v)
        case int, int64, float64:
            replacement = fmt.Sprintf("%v", v)
        default:
            replacement = "null"
        }

        expr = strings.ReplaceAll(expr, placeholder, replacement)
    }

    // TODO: Use a proper expression evaluator library
    // For now, implement basic evaluation

    return true, nil
}

// getNestedValue retrieves a nested value from a map using dot notation
func (dc *DependencyChecker) getNestedValue(
    values map[string]interface{},
    key string,
) interface{} {
    parts := strings.Split(key, ".")
    current := values

    for i, part := range parts {
        if i == len(parts)-1 {
            return current[part]
        }

        next, ok := current[part].(map[string]interface{})
        if !ok {
            return nil
        }
        current = next
    }

    return nil
}
```

**파일 경로**:
- `backend/internal/services/dependency_checker.go` (신규)
- `backend/internal/services/dependency_checker_test.go` (신규)

**검증 기준**:
- [ ] requires 규칙이 올바르게 검증됨
- [ ] conflicts 규칙이 올바르게 검증됨
- [ ] 중첩된 필드 참조가 동작함
- [ ] 모든 단위 테스트 통과

---

### Task 3.2: 의존성 검증 API (완료)

**목적**: 프론트엔드에서 실시간 검증을 위한 API 엔드포인트를 구현합니다.

**구현 내용**:
```go
// backend/internal/handlers/container_configuration.go에 추가

// ValidateConfiguration validates configuration values
func (h *ContainerConfigurationHandler) ValidateConfiguration(c *gin.Context) {
    serviceID, _ := strconv.ParseUint(c.Param("service_id"), 10, 64)
    containerID, _ := strconv.ParseUint(c.Param("container_id"), 10, 64)

    var req struct {
        Values map[string]interface{} `json:"values"`
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Load container configuration
    var serviceContainer models.ServiceContainer
    if err := h.db.Where("service_id = ? AND container_id = ?", serviceID, containerID).
        Preload("ContainerVersion").
        First(&serviceContainer).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Service container not found"})
        return
    }

    var config models.ContainerConfiguration
    if err := h.db.Where("container_version_id = ?", serviceContainer.ContainerVersionID).
        First(&config).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Container configuration not found"})
        return
    }

    // Parse dependency rules
    var rules []services.DependencyRule
    if err := json.Unmarshal(config.DependencyRules, &rules); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse dependency rules"})
        return
    }

    // Validate
    checker := services.NewDependencyChecker()
    errors := checker.ValidateConfiguration(rules, req.Values)

    c.JSON(http.StatusOK, gin.H{
        "valid":  len(errors) == 0,
        "errors": errors,
    })
}
```

**라우트 등록**:
```go
// backend/internal/server/server.go에 추가
api.POST("/services/:service_id/containers/:container_id/validate",
    configHandler.ValidateConfiguration)
```

**검증 기준**:
- [ ] POST /api/v1/services/{id}/containers/{id}/validate 동작
- [ ] 올바른 검증 결과 반환
- [ ] 오류 메시지가 명확함

---

### Task 3.3: 프론트엔드 실시간 검증

**목적**: 사용자가 설정을 변경할 때마다 실시간으로 의존성을 검증합니다.

**구현 내용**:
```typescript
// frontend/src/hooks/useDependencyValidation.ts
import { useState, useEffect, useCallback } from 'react'
import { ConfigurationValues } from '@/types/configuration'
import { api } from '@/services/api'
import { debounce } from 'lodash'

interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
}

interface ValidationError {
  field: string
  message: string
  rule: string
}

export const useDependencyValidation = (
  serviceId: string,
  containerId: string,
  values: ConfigurationValues
) => {
  const [validationResult, setValidationResult] = useState<ValidationResult>({
    valid: true,
    errors: []
  })
  const [isValidating, setIsValidating] = useState(false)

  const validateValues = useCallback(
    debounce(async (vals: ConfigurationValues) => {
      setIsValidating(true)

      try {
        const response = await api.post(
          `/services/${serviceId}/containers/${containerId}/validate`,
          { values: vals }
        )

        setValidationResult(response.data)
      } catch (error) {
        console.error('Validation failed:', error)
      } finally {
        setIsValidating(false)
      }
    }, 500),
    [serviceId, containerId]
  )

  useEffect(() => {
    validateValues(values)
  }, [values, validateValues])

  return { validationResult, isValidating }
}
```

**ConfigurationForm 수정**:
```typescript
// frontend/src/components/configuration/ConfigurationForm.tsx 수정
import { useDependencyValidation } from '@/hooks/useDependencyValidation'

export const ConfigurationForm: React.FC<ConfigurationFormProps> = ({
  schema,
  serviceId,
  containerId,
  initialValues = {},
  onChange
}) => {
  const [values, setValues] = useState<ConfigurationValues>(initialValues)
  const { validationResult, isValidating } = useDependencyValidation(
    serviceId,
    containerId,
    values
  )

  // ... rest of the component

  return (
    <div className="configuration-form space-y-6">
      {isValidating && (
        <div className="validation-indicator text-sm text-gray-500">
          검증 중...
        </div>
      )}

      {!validationResult.valid && (
        <div className="validation-errors bg-red-50 border border-red-200 rounded p-4">
          <h4 className="font-semibold text-red-800 mb-2">설정 오류</h4>
          <ul className="list-disc list-inside space-y-1">
            {validationResult.errors.map((error, idx) => (
              <li key={idx} className="text-sm text-red-700">
                {error.message}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* ... rest of the form */}
    </div>
  )
}
```

**검증 기준**:
- [ ] 설정 변경 시 자동으로 검증 실행
- [ ] 검증 결과가 실시간으로 표시됨
- [ ] 디바운싱이 정상 동작 (과도한 API 호출 방지)
- [ ] 오류 메시지가 사용자 친화적

---

## Phase 4: 빌드 프로세스 통합

**목표**: 템플릿 렌더링과 에셋 해결을 빌드 프로세스에 통합합니다.

**예상 기간**: 2주

### Task 4.1: 빌드 프로세스 확장

**목적**: 기존 빌드 프로세스에 템플릿 렌더링 단계를 추가합니다.

**구현 내용**:
```go
// backend/internal/services/build_service.go (기존 파일 확장)
package services

import (
    "context"
    "encoding/json"
    "fmt"
    "path/filepath"

    "burndler/internal/models"
    "burndler/internal/storage"
)

type BuildService struct {
    db                *gorm.DB
    storage           storage.Interface
    templateEngine    *TemplateEngine
    dependencyChecker *DependencyChecker
    mergerService     *MergerService
    linterService     *LinterService
    packagerService   *PackagerService
}

func NewBuildService(
    db *gorm.DB,
    storage storage.Interface,
) *BuildService {
    return &BuildService{
        db:                db,
        storage:           storage,
        templateEngine:    NewTemplateEngine(),
        dependencyChecker: NewDependencyChecker(),
        // ... initialize other services
    }
}

type BuildContext struct {
    Build             *models.Build
    Service           *models.Service
    Configurations    map[uint]*models.ContainerConfiguration
    ResolvedVariables map[uint]map[string]interface{}
    RenderedFiles     map[string]string
    TempDirectory     string
}

// ExecuteBuild executes the full build pipeline
func (bs *BuildService) ExecuteBuild(ctx context.Context, buildID string) error {
    // Load build
    var build models.Build
    if err := bs.db.Preload("Service.ServiceContainers.ContainerVersion").
        Where("id = ?", buildID).
        First(&build).Error; err != nil {
        return fmt.Errorf("failed to load build: %w", err)
    }

    buildCtx := &BuildContext{
        Build:             &build,
        Service:           build.Service,
        Configurations:    make(map[uint]*models.ContainerConfiguration),
        ResolvedVariables: make(map[uint]map[string]interface{}),
        RenderedFiles:     make(map[string]string),
    }

    // Execute build stages
    stages := []struct {
        name string
        fn   func(context.Context, *BuildContext) error
    }{
        {"validation", bs.validateConfiguration},
        {"configuration", bs.resolveConfiguration},
        {"template_render", bs.renderTemplates},
        {"asset_resolution", bs.resolveAssets},
        {"compose_merge", bs.mergeCompose},
        {"linting", bs.lintCompose},
        {"packaging", bs.packageInstaller},
    }

    for _, stage := range stages {
        build.Status = fmt.Sprintf("building:%s", stage.name)
        if err := bs.updateBuildStatus(&build); err != nil {
            return err
        }

        if err := stage.fn(ctx, buildCtx); err != nil {
            build.Status = "failed"
            build.Error = err.Error()
            bs.updateBuildStatus(&build)
            return fmt.Errorf("stage %s failed: %w", stage.name, err)
        }
    }

    build.Status = "completed"
    return bs.updateBuildStatus(&build)
}

// resolveConfiguration loads and validates all configurations
func (bs *BuildService) resolveConfiguration(ctx context.Context, buildCtx *BuildContext) error {
    for _, sc := range buildCtx.Service.ServiceContainers {
        if !sc.Enabled {
            continue
        }

        // Load container configuration
        var config models.ContainerConfiguration
        if err := bs.db.Where("container_version_id = ?", sc.ContainerVersionID).
            Preload("Files").
            Preload("Assets").
            First(&config).Error; err != nil {
            // No configuration defined, skip
            continue
        }

        buildCtx.Configurations[sc.ContainerID] = &config

        // Resolve variables
        variables := bs.resolveVariables(buildCtx.Service, &sc, &config)
        buildCtx.ResolvedVariables[sc.ContainerID] = variables

        // Validate dependencies
        var rules []DependencyRule
        if err := json.Unmarshal(config.DependencyRules, &rules); err == nil {
            errors := bs.dependencyChecker.ValidateConfiguration(rules, variables)
            if len(errors) > 0 {
                return fmt.Errorf("dependency validation failed for container %d: %v", sc.ContainerID, errors)
            }
        }
    }

    return nil
}

// renderTemplates renders all template files
func (bs *BuildService) renderTemplates(ctx context.Context, buildCtx *BuildContext) error {
    for containerID, config := range buildCtx.Configurations {
        variables := buildCtx.ResolvedVariables[containerID]

        // Render template files
        for _, file := range config.Files {
            if file.FileType != "template" {
                continue
            }

            // Load template content from storage
            content, err := bs.loadFileContent(file.StoragePath)
            if err != nil {
                return fmt.Errorf("failed to load template %s: %w", file.FilePath, err)
            }

            // Render template
            rendered, err := bs.templateEngine.Render(
                string(content),
                file.TemplateFormat,
                variables,
            )
            if err != nil {
                return fmt.Errorf("failed to render template %s: %w", file.FilePath, err)
            }

            // Store rendered content
            namespacedPath := bs.applyNamespace(file.FilePath, containerID, buildCtx.Service)
            buildCtx.RenderedFiles[namespacedPath] = rendered
        }
    }

    return nil
}

// resolveVariables resolves variables with proper precedence
func (bs *BuildService) resolveVariables(
    service *models.Service,
    serviceContainer *models.ServiceContainer,
    config *models.ContainerConfiguration,
) map[string]interface{} {
    variables := make(map[string]interface{})

    // Global variables
    variables["SERVICE_NAME"] = service.Name
    variables["SERVICE_ID"] = service.ID

    // Service variables
    if service.Variables != nil {
        var serviceVars map[string]interface{}
        json.Unmarshal(service.Variables, &serviceVars)
        for k, v := range serviceVars {
            variables[k] = v
        }
    }

    // Container overrides
    effectiveVars := serviceContainer.GetEffectiveVariables()
    for k, v := range effectiveVars {
        variables[k] = v
    }

    return variables
}

// applyNamespace applies namespace prefix to file path
func (bs *BuildService) applyNamespace(
    filePath string,
    containerID uint,
    service *models.Service,
) string {
    var container models.Container
    bs.db.First(&container, containerID)

    namespace := fmt.Sprintf("%s_%d", service.Name, service.ID)
    return filepath.Join(namespace, container.Name, filePath)
}

// loadFileContent loads file content from storage
func (bs *BuildService) loadFileContent(storagePath string) ([]byte, error) {
    return bs.storage.Retrieve(storagePath)
}

// updateBuildStatus updates build status in database
func (bs *BuildService) updateBuildStatus(build *models.Build) error {
    return bs.db.Save(build).Error
}
```

**파일 경로**:
- `backend/internal/services/build_service.go` (확장)

**검증 기준**:
- [ ] 빌드 프로세스가 새로운 단계를 포함함
- [ ] 템플릿 렌더링이 정상 동작함
- [ ] 변수 해결이 올바른 우선순위로 동작함
- [ ] 기존 빌드 기능에 영향 없음

---

### Task 4.2: 에셋 해결 및 패키징

**목적**: 조건부 에셋을 해결하고 인스톨러에 포함시킵니다.

**구현 내용**:
```go
// backend/internal/services/build_service.go에 추가

// resolveAssets resolves which assets to include
func (bs *BuildService) resolveAssets(ctx context.Context, buildCtx *BuildContext) error {
    for containerID, config := range buildCtx.Configurations {
        variables := buildCtx.ResolvedVariables[containerID]

        for _, asset := range config.Assets {
            // Evaluate include condition
            if asset.IncludeCondition != "" {
                include, err := bs.evaluateCondition(asset.IncludeCondition, variables)
                if err != nil {
                    return fmt.Errorf("failed to evaluate asset condition %s: %w", asset.IncludeCondition, err)
                }
                if !include {
                    continue // Skip this asset
                }
            }

            // Handle based on storage type
            switch asset.StorageType {
            case "embedded":
                // Load asset content
                content, err := bs.storage.Retrieve(asset.StoragePath)
                if err != nil {
                    return fmt.Errorf("failed to load embedded asset %s: %w", asset.FilePath, err)
                }

                // Store in rendered files
                namespacedPath := bs.applyNamespace(asset.FilePath, containerID, buildCtx.Service)
                buildCtx.RenderedFiles[namespacedPath] = string(content)

            case "download":
                // Generate download URL
                downloadURL, err := bs.generateDownloadURL(asset.StoragePath)
                if err != nil {
                    return fmt.Errorf("failed to generate download URL for %s: %w", asset.FilePath, err)
                }

                // Add to download manifest
                // This will be handled in packaging stage
            }
        }
    }

    return nil
}

// evaluateCondition evaluates a simple boolean condition
func (bs *BuildService) evaluateCondition(
    condition string,
    variables map[string]interface{},
) (bool, error) {
    // Use dependency checker's evaluation logic
    return bs.dependencyChecker.evaluateCondition(condition, variables)
}

// generateDownloadURL generates a signed download URL
func (bs *BuildService) generateDownloadURL(storagePath string) (string, error) {
    // Implementation depends on storage backend
    // For S3, generate presigned URL
    // For local FS, generate API endpoint

    return fmt.Sprintf("/api/v1/assets/download?path=%s", storagePath), nil
}
```

**검증 기준**:
- [ ] 조건부 에셋이 올바르게 포함/제외됨
- [ ] 임베디드 에셋이 인스톨러에 포함됨
- [ ] 다운로드 에셋의 URL이 생성됨

---

## Phase 5: 고급 기능 및 최적화

**목표**: 사용자 경험을 개선하고 성능을 최적화합니다.

**예상 기간**: 2-3주

### Task 5.1: 템플릿 함수 확장

**목적**: 템플릿에서 사용할 수 있는 고급 함수를 추가합니다.

**구현 내용**:
```go
// backend/internal/services/template_functions.go (신규)
package services

import (
    "crypto/rand"
    "crypto/sha256"
    "encoding/base64"
    "encoding/hex"
    "fmt"
    "math/big"
    "net"
    "os"
    "text/template"
    "time"

    "github.com/google/uuid"
)

// GetExtendedTemplateFuncMap returns extended template functions
func GetExtendedTemplateFuncMap() template.FuncMap {
    return template.FuncMap{
        // String functions
        "upper":     strings.ToUpper,
        "lower":     strings.ToLower,
        "trim":      strings.TrimSpace,
        "replace":   strings.ReplaceAll,
        "contains":  strings.Contains,
        "hasPrefix": strings.HasPrefix,
        "hasSuffix": strings.HasSuffix,
        "split":     strings.Split,
        "join":      strings.Join,

        // Math functions
        "add": func(a, b int) int { return a + b },
        "sub": func(a, b int) int { return a - b },
        "mul": func(a, b int) int { return a * b },
        "div": func(a, b int) int { return a / b },

        // Conditional functions
        "default": func(defaultVal, val interface{}) interface{} {
            if val == nil || val == "" {
                return defaultVal
            }
            return val
        },
        "eq": func(a, b interface{}) bool { return a == b },
        "ne": func(a, b interface{}) bool { return a != b },

        // Utility functions
        "env": os.Getenv,
        "uuid": func() string {
            return uuid.New().String()
        },
        "timestamp": func() int64 {
            return time.Now().Unix()
        },

        // Security functions
        "generatePassword": generateSecurePassword,
        "hash":            hashString,
        "base64encode":    base64Encode,
        "base64decode":    base64Decode,

        // Network functions
        "randomPort": generateRandomPort,
        "localIP":    getLocalIP,
    }
}

// generateSecurePassword generates a random password
func generateSecurePassword(length int) string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()"
    password := make([]byte, length)

    for i := range password {
        n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
        password[i] = charset[n.Int64()]
    }

    return string(password)
}

// hashString returns SHA256 hash of a string
func hashString(s string) string {
    hash := sha256.Sum256([]byte(s))
    return hex.EncodeToString(hash[:])
}

// base64Encode encodes a string to base64
func base64Encode(s string) string {
    return base64.StdEncoding.EncodeToString([]byte(s))
}

// base64Decode decodes a base64 string
func base64Decode(s string) (string, error) {
    decoded, err := base64.StdEncoding.DecodeString(s)
    if err != nil {
        return "", err
    }
    return string(decoded), nil
}

// generateRandomPort generates a random port in a range
func generateRandomPort(min, max int) int {
    n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min+1)))
    return min + int(n.Int64())
}

// getLocalIP returns the local IP address
func getLocalIP() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return "127.0.0.1"
    }

    for _, addr := range addrs {
        if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }

    return "127.0.0.1"
}
```

**TemplateEngine 수정**:
```go
// backend/internal/services/template_engine.go 수정
func NewTemplateEngine() *TemplateEngine {
    return &TemplateEngine{
        funcMap: GetExtendedTemplateFuncMap(),
    }
}
```

**검증 기준**:
- [ ] 모든 새 함수가 템플릿에서 동작함
- [ ] 보안 관련 함수가 안전하게 동작함
- [ ] 단위 테스트 통과

---

### Task 5.2: 설정 내보내기/가져오기

**목적**: 설정을 JSON 파일로 내보내거나 가져올 수 있게 합니다.

**구현 내용**:
```go
// backend/internal/handlers/service_configuration.go (신규)
package handlers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "burndler/internal/models"
    "gorm.io/gorm"
)

type ServiceConfigurationHandler struct {
    db *gorm.DB
}

func NewServiceConfigurationHandler(db *gorm.DB) *ServiceConfigurationHandler {
    return &ServiceConfigurationHandler{db: db}
}

// ExportConfiguration exports all service configurations as JSON
func (h *ServiceConfigurationHandler) ExportConfiguration(c *gin.Context) {
    serviceID, _ := strconv.ParseUint(c.Param("service_id"), 10, 64)

    var configs []models.ServiceConfiguration
    if err := h.db.Where("service_id = ?", serviceID).
        Preload("Container").
        Find(&configs).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load configurations"})
        return
    }

    // Build export structure
    export := make(map[string]interface{})
    export["version"] = "1.0"
    export["service_id"] = serviceID

    containerConfigs := make(map[string]interface{})
    for _, config := range configs {
        var values map[string]interface{}
        json.Unmarshal(config.ConfigurationValues, &values)
        containerConfigs[config.Container.Name] = values
    }
    export["containers"] = containerConfigs

    c.JSON(http.StatusOK, export)
}

// ImportConfiguration imports configurations from JSON
func (h *ServiceConfigurationHandler) ImportConfiguration(c *gin.Context) {
    serviceID, _ := strconv.ParseUint(c.Param("service_id"), 10, 64)

    var importData struct {
        Version    string                            `json:"version"`
        Containers map[string]map[string]interface{} `json:"containers"`
    }

    if err := c.ShouldBindJSON(&importData); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // Validate version
    if importData.Version != "1.0" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported export version"})
        return
    }

    // Import configurations
    for containerName, values := range importData.Containers {
        // Find container by name
        var container models.Container
        if err := h.db.Where("name = ?", containerName).First(&container).Error; err != nil {
            continue // Skip unknown containers
        }

        // Update or create service configuration
        valuesJSON, _ := json.Marshal(values)
        config := &models.ServiceConfiguration{
            ServiceID:           uint(serviceID),
            ContainerID:         container.ID,
            ConfigurationValues: valuesJSON,
        }

        h.db.Where("service_id = ? AND container_id = ?", serviceID, container.ID).
            Assign(config).
            FirstOrCreate(config)
    }

    c.JSON(http.StatusOK, gin.H{"message": "Configuration imported successfully"})
}
```

**프론트엔드 구현**:
```typescript
// frontend/src/components/configuration/ConfigurationActions.tsx
import React from 'react'
import { api } from '@/services/api'

interface ConfigurationActionsProps {
  serviceId: string
}

export const ConfigurationActions: React.FC<ConfigurationActionsProps> = ({
  serviceId
}) => {
  const handleExport = async () => {
    try {
      const response = await api.get(`/services/${serviceId}/configuration/export`)
      const blob = new Blob([JSON.stringify(response.data, null, 2)], {
        type: 'application/json'
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `service-${serviceId}-config.json`
      a.click()
      URL.revokeObjectURL(url)
    } catch (error) {
      console.error('Export failed:', error)
      alert('설정 내보내기에 실패했습니다')
    }
  }

  const handleImport = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0]
    if (!file) return

    try {
      const text = await file.text()
      const data = JSON.parse(text)

      await api.post(`/services/${serviceId}/configuration/import`, data)
      alert('설정을 가져왔습니다')
      window.location.reload()
    } catch (error) {
      console.error('Import failed:', error)
      alert('설정 가져오기에 실패했습니다')
    }
  }

  return (
    <div className="configuration-actions flex space-x-3">
      <button
        onClick={handleExport}
        className="btn btn-secondary px-4 py-2 border rounded hover:bg-gray-100"
      >
        📥 내보내기
      </button>

      <label className="btn btn-secondary px-4 py-2 border rounded hover:bg-gray-100 cursor-pointer">
        📤 가져오기
        <input
          type="file"
          accept=".json"
          onChange={handleImport}
          className="hidden"
        />
      </label>
    </div>
  )
}
```

**검증 기준**:
- [ ] 설정을 JSON으로 내보낼 수 있음
- [ ] JSON 파일을 가져와서 설정 복원 가능
- [ ] 버전 호환성 체크가 동작함

---

## 검증 및 테스트

### 통합 테스트 시나리오

**시나리오 1: 기본 템플릿 워크플로우**
```bash
# 1. Container 및 Version 생성
POST /api/v1/containers
POST /api/v1/containers/1/versions

# 2. Configuration 생성
POST /api/v1/containers/1/versions/1/configuration
{
  "ui_schema": {
    "sections": [
      {
        "id": "database",
        "title": "데이터베이스 설정",
        "fields": [
          {
            "key": "Database.Host",
            "type": "string",
            "label": "호스트",
            "required": true
          }
        ]
      }
    ]
  },
  "dependency_rules": []
}

# 3. 템플릿 파일 업로드
POST /api/v1/containers/1/versions/1/files
{
  "file_path": "config/database.yaml",
  "file_type": "template",
  "template_format": "yaml",
  "content": "host: {{ .Database.Host }}"
}

# 4. Service 생성 및 Container 추가
POST /api/v1/services
POST /api/v1/services/1/containers

# 5. Service Configuration 설정
PUT /api/v1/services/1/containers/1/config
{
  "configuration_values": {
    "Database.Host": "localhost"
  }
}

# 6. Build 실행
POST /api/v1/services/1/build

# 7. 검증: 렌더링된 파일 확인
GET /api/v1/builds/{build_id}/download
```

**시나리오 2: 의존성 검증**
```bash
# Configuration with dependency rules
POST /api/v1/containers/1/versions/1/configuration
{
  "ui_schema": {...},
  "dependency_rules": [
    {
      "type": "requires",
      "field": "SSL.Enabled",
      "condition": "{{.SSL.Enabled}} == true",
      "target": "SSL.CertificatePath",
      "message": "SSL이 활성화되면 인증서 경로가 필요합니다"
    }
  ]
}

# 검증 테스트
POST /api/v1/services/1/containers/1/validate
{
  "values": {
    "SSL.Enabled": true,
    "SSL.CertificatePath": ""  // Should fail validation
  }
}
```

**시나리오 3: 조건부 파일**
```bash
# 조건부 템플릿 파일
POST /api/v1/containers/1/versions/1/files
{
  "file_path": "config/cache.yaml",
  "file_type": "template",
  "display_condition": "{{.Cache.Enabled}} == true"
}

# Build 시 Cache.Enabled = false이면 파일이 제외됨
```

### 성능 테스트

**테스트 항목**:
- [ ] 10개 Container, 각 20개 템플릿 파일 빌드 시간 < 2분
- [ ] 실시간 검증 응답 시간 < 500ms
- [ ] 파일 구조 해결 시간 < 200ms
- [ ] 1GB 에셋 파일 처리 가능

### 사용자 시나리오 테스트

**개발자 워크플로우**:
1. Container 생성
2. Configuration 정의 (UI 스키마 + 의존성 규칙)
3. 템플릿 파일 작성 및 업로드
4. 에셋 파일 업로드 (조건 설정)
5. 테스트 빌드 실행

**현장 엔지니어 워크플로우**:
1. Service 생성
2. Container 선택
3. GUI로 설정 변경
4. 파일 구조 미리보기 확인
5. 의존성 오류 해결
6. Build 실행
7. 인스톨러 다운로드

---

## 마이그레이션 가이드

기존 Container들을 템플릿 시스템으로 마이그레이션하는 방법:

### Step 1: 기존 Container 분석
- 현재 docker-compose.yaml 분석
- 설정 파일들 식별
- 변수로 만들 값들 결정

### Step 2: UI 스키마 작성
- 사용자가 변경할 수 있어야 하는 값들을 필드로 정의
- 섹션으로 그룹화
- 의존성 규칙 정의

### Step 3: 템플릿 생성
- 기존 파일을 템플릿으로 변환
- 하드코딩된 값을 변수로 치환
- 조건부 블록 추가

### Step 4: 검증 및 테스트
- 다양한 설정 조합으로 빌드 테스트
- 의존성 검증 테스트
- 인스톨러 실행 테스트

---

## 문제 해결 가이드

### 템플릿 렌더링 오류
**증상**: "template execution error"
**원인**: 변수가 정의되지 않음
**해결**: `default` 함수 사용 또는 변수 초기화

### 의존성 검증 실패
**증상**: "dependency validation failed"
**원인**: 필수 필드가 비어있거나 충돌
**해결**: UI에서 오류 메시지 확인 후 필드 수정

### 파일 구조 표시 안됨
**증상**: 빈 파일 트리
**원인**: 조건 평가 오류
**해결**: 브라우저 콘솔 확인, 조건 표현식 수정

---

## 참고 자료

### Go Template 문법
- https://pkg.go.dev/text/template
- 기본 문법: `{{ .Variable }}`
- 조건문: `{{if .Condition}}...{{end}}`
- 반복문: `{{range .Items}}...{{end}}`

### YAML 모범 사례
- 들여쓰기는 공백 2칸
- 문자열 따옴표는 특수문자 있을 때만
- 앵커(&)와 참조(*)로 중복 제거

### JSON Schema
- UI 스키마 구조 정의
- 검증 규칙 표준화

---

## 다음 단계

이 문서의 각 Task를 순서대로 구현하세요. 각 Task는:
1. 명확한 목적
2. 구체적인 구현 내용
3. 파일 경로
4. 검증 기준

을 포함하고 있어 독립적으로 명령하고 실행할 수 있습니다.

각 Task 완료 후:
- [ ] 코드 리뷰
- [ ] 테스트 실행
- [ ] 문서 업데이트
- [ ] 다음 Task 진행