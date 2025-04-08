package resources

// PromptBuilder is a builder for creating prompts
type PromptBuilder struct {
	prompt Prompt
}

// PromptMessageBuilder is a builder for creating prompt messages
type PromptMessageBuilder struct {
	message PromptMessage
	prompt  *PromptBuilder
}

// PromptArgumentBuilder is a builder for creating prompt arguments
type PromptArgumentBuilder struct {
	argument PromptArgument
	prompt   *PromptBuilder
}

// NewPrompt creates a new prompt builder
func NewPrompt(name string) *PromptBuilder {
	return &PromptBuilder{
		prompt: Prompt{
			Name:      name,
			Arguments: []PromptArgument{},
			Messages:  []PromptMessage{},
		},
	}
}

// WithDescription sets the description of the prompt
func (b *PromptBuilder) WithDescription(description string) *PromptBuilder {
	b.prompt.Description = description
	return b
}

// WithArgument adds an argument to the prompt
func (b *PromptBuilder) WithArgument(name string) *PromptArgumentBuilder {
	return &PromptArgumentBuilder{
		argument: PromptArgument{
			Name: name,
		},
		prompt: b,
	}
}

// WithUserMessage adds a user message to the prompt
func (b *PromptBuilder) WithUserMessage(content string) *PromptBuilder {
	b.prompt.Messages = append(b.prompt.Messages, PromptMessage{
		Role:    "user",
		Content: content,
	})
	return b
}

// WithAssistantMessage adds an assistant message to the prompt
func (b *PromptBuilder) WithAssistantMessage(content string) *PromptBuilder {
	b.prompt.Messages = append(b.prompt.Messages, PromptMessage{
		Role:    "assistant",
		Content: content,
	})
	return b
}

// WithMessage adds a custom message to the prompt
func (b *PromptBuilder) WithMessage() *PromptMessageBuilder {
	return &PromptMessageBuilder{
		message: PromptMessage{},
		prompt:  b,
	}
}

// Build builds the prompt
func (b *PromptBuilder) Build() Prompt {
	return b.prompt
}

// Required marks the argument as required
func (b *PromptArgumentBuilder) Required() *PromptArgumentBuilder {
	b.argument.Required = true
	return b
}

// Description sets the description of the argument
func (b *PromptArgumentBuilder) Description(description string) *PromptArgumentBuilder {
	b.argument.Description = description
	return b
}

// Add adds the argument to the prompt and returns the prompt builder
func (b *PromptArgumentBuilder) Add() *PromptBuilder {
	b.prompt.prompt.Arguments = append(b.prompt.prompt.Arguments, b.argument)
	return b.prompt
}

// Role sets the role of the message
func (b *PromptMessageBuilder) Role(role string) *PromptMessageBuilder {
	b.message.Role = role
	return b
}

// Content sets the content of the message
func (b *PromptMessageBuilder) Content(content interface{}) *PromptMessageBuilder {
	b.message.Content = content
	return b
}

// Add adds the message to the prompt and returns the prompt builder
func (b *PromptMessageBuilder) Add() *PromptBuilder {
	b.prompt.prompt.Messages = append(b.prompt.prompt.Messages, b.message)
	return b.prompt
}
