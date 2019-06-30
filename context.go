package wish

// A Provider is a function that will return a Context
// to be used for handling a server request.
// A Provider must always return values of the same
// concrete type.
type Provider func() (Context, error)

// Context represents a type that can be used to handle a
// server request.
type Context interface {
	context()
}

// ContextImpl must be embedded in all types that implement Context.
// It provides several utility methods for use in handlers.
type ContextImpl struct{}

func (*ContextImpl) context() {}
