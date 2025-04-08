package resources

// ToolBuilder is a builder for creating resources
type ToolBuilder struct {
	tool Tool
}

// ParameterBuilder is a builder for creating tool parameters
type ParameterBuilder struct {
	name     string
	property SchemaProperty
	tool     *ToolBuilder
}

// NewTool creates a new tool builder
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		tool: Tool{
			Name: name,
			InputSchema: InputSchema{
				Type:       "object",
				Properties: make(map[string]SchemaProperty),
				Required:   []string{},
			},
		},
	}
}

// WithDescription sets the description of the tool
func (b *ToolBuilder) WithDescription(description string) *ToolBuilder {
	b.tool.Description = description
	return b
}

// WithString adds a string parameter to the tool
func (b *ToolBuilder) WithString(name string) *ParameterBuilder {
	return &ParameterBuilder{
		name: name,
		property: SchemaProperty{
			Type: "string",
		},
		tool: b,
	}
}

// WithInteger adds an integer parameter to the tool
func (b *ToolBuilder) WithInteger(name string) *ParameterBuilder {
	return &ParameterBuilder{
		name: name,
		property: SchemaProperty{
			Type: "integer",
		},
		tool: b,
	}
}

// WithBoolean adds a boolean parameter to the tool
func (b *ToolBuilder) WithBoolean(name string) *ParameterBuilder {
	return &ParameterBuilder{
		name: name,
		property: SchemaProperty{
			Type: "boolean",
		},
		tool: b,
	}
}

// WithObject adds an object parameter to the tool
func (b *ToolBuilder) WithObject(name string) *ParameterBuilder {
	return &ParameterBuilder{
		name: name,
		property: SchemaProperty{
			Type: "object",
		},
		tool: b,
	}
}

// WithArray adds an array parameter to the tool
func (b *ToolBuilder) WithArray(name string) *ParameterBuilder {
	return &ParameterBuilder{
		name: name,
		property: SchemaProperty{
			Type: "array",
		},
		tool: b,
	}
}

// Build builds the tool
func (b *ToolBuilder) Build() Tool {
	return b.tool
}

// Required marks the parameter as required
func (b *ParameterBuilder) Required() *ParameterBuilder {
	b.tool.tool.InputSchema.Required = append(b.tool.tool.InputSchema.Required, b.name)
	return b
}

// Description sets the description of the parameter
func (b *ParameterBuilder) Description(description string) *ParameterBuilder {
	b.property.Description = description
	return b
}

// Default sets the default value of the parameter
func (b *ParameterBuilder) Default(value interface{}) *ParameterBuilder {
	b.property.Default = value
	return b
}

// Add adds the parameter to the tool and returns the tool builder
func (b *ParameterBuilder) Add() *ToolBuilder {
	b.tool.tool.InputSchema.Properties[b.name] = b.property
	return b.tool
}
