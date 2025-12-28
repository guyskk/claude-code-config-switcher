# core-config 规范变更

## ADDED Requirements

### Requirement: 类型安全的配置结构

系统 SHALL 使用强类型的 Go 结构体来表示配置，而不是使用 `map[string]interface{}`。

#### Scenario: Env 类型
- **GIVEN** Env 定义为 `map[string]string`
- **WHEN** 序列化/反序列化 JSON
- **THEN** 应当正确处理环境变量映射

#### Scenario: ProviderConfig 类型
- **GIVEN** ProviderConfig 包含可选的 Env 字段
- **WHEN** 反序列化提供商配置
- **THEN** 应当正确解析提供商的环境变量

#### Scenario: Settings 类型
- **GIVEN** Settings 包含 permissions、alwaysThinkingEnabled 和 env 字段
- **WHEN** 反序列化设置
- **THEN** 应当正确解析所有字段

#### Scenario: Config 类型
- **GIVEN** Config 包含 settings、current_provider 和 providers 字段
- **WHEN** 反序列化完整配置
- **THEN** 应当正确解析整个配置结构

### Requirement: 配置加载

系统 SHALL 提供从文件加载配置的功能，并返回类型安全的配置对象。

#### Scenario: 加载有效配置
- **GIVEN** 存在有效的 ccc.json 文件
- **WHEN** 调用 `Load()` 函数
- **THEN** 应当返回 Config 对象
- **AND** error 应当为 nil

#### Scenario: 配置文件不存在
- **GIVEN** ccc.json 文件不存在
- **WHEN** 调用 `Load()` 函数
- **THEN** 应当返回错误
- **AND** 错误信息应当包含 "failed to read config file"

#### Scenario: 配置格式错误
- **GIVEN** ccc.json 包含无效的 JSON
- **WHEN** 调用 `Load()` 函数
- **THEN** 应当返回错误
- **AND** 错误信息应当包含 "failed to parse config file"

### Requirement: 配置保存

系统 SHALL 提供将配置保存到文件的功能。

#### Scenario: 保存配置
- **GIVEN** 一个有效的 Config 对象
- **WHEN** 调用 `Save()` 函数
- **THEN** 应当创建或更新 ccc.json 文件
- **AND** 文件内容应当是格式化的 JSON（缩进 2 个空格）
- **AND** 文件权限应当为 0644

#### Scenario: 目录不存在
- **GIVEN** 配置目录不存在
- **WHEN** 调用 `Save()` 函数
- **THEN** 应当自动创建目录
- **AND** 目录权限应当为 0755

### Requirement: 配置路径解析

系统 SHALL 提供获取配置文件路径的函数。

#### Scenario: 默认配置路径
- **WHEN** 未设置 CCC_CONFIG_DIR 环境变量
- **THEN** 配置路径应当为 `~/.claude/ccc.json`

#### Scenario: 自定义配置路径
- **GIVEN** CCC_CONFIG_DIR 设置为 `/custom/path`
- **WHEN** 获取配置路径
- **THEN** 配置路径应当为 `/custom/path/ccc.json`
