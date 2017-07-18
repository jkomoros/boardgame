/*

errors is a package that implements FriendlyError. FriendlyError implements
the error interface, but also has a FriendlyError() method that returns a
message that is reasonable to show to end-users.

Every method that returns a FriendlyError also accepts 0 to n Fields objects
may be provided and they will be combined into a single Fields, with later
values overwriting newer ones.

FriendlyErrors also have With* methods that return a copy of the error with
the given changes. This allows a pattern where a method defines a default
FriendlyError with the friendly error message to use and then when it needs to
return an error uses WithError() to add the specific error message.

*/
package errors

const DefaultFriendlyError = "An error occurred"

type Fields map[string]interface{}

type Friendly struct {
	msg         string
	friendlyMsg string
	secureMsg   string
	fields      Fields
}

//New creates a new errors.Friendly with the given msg.
func New(msg string, fields ...Fields) *Friendly {
	return &Friendly{
		msg:    msg,
		fields: combineFields(fields...),
	}
}

//NewFriendly returns a new errors.Friendly with the given FriendlyError.
func NewFriendly(friendlyMsg string, fields ...Fields) *Friendly {
	return &Friendly{
		friendlyMsg: friendlyMsg,
		fields:      combineFields(fields...),
	}
}

func combineFields(fields ...Fields) Fields {
	var result = make(Fields)
	for _, field := range fields {
		for key, val := range field {
			result[key] = val
		}
	}
	return result
}

//SecureError returns the error message that should only be shown in secure
//contexts, because it may include secret information. If no SecureError
//message has been provided, will return Error()
func (f *Friendly) SecureError() string {
	if f.secureMsg == "" {
		return f.Error()
	}
	return f.secureMsg
}

//Error returns the error message, implementing the error interface. If no
//specific message has been provided, will fall back on FriendlyError(). The
//Error() value is OK to show in insecure contexts (i.e. on the client) but it
//just might be confusing to users.
func (f *Friendly) Error() string {
	if f.msg == "" {
		return f.FriendlyError()
	}
	return f.msg
}

//FriendlyError is the error string that is OK to show in insecure contexts to
//end-users. It is generally a much simplified version of the message. If no
//specific FriendlyMessage has been provided, will return DefaultFriendlyError.
func (f *Friendly) FriendlyError() string {
	if f.friendlyMsg == "" {
		return DefaultFriendlyError
	}
	return f.friendlyMsg
}

//Fields returns the Fields object for this error.
func (f *Friendly) Fields() Fields {
	return f.fields
}

//Extend returns a new FriendlyError where the Error() message is prepended
//with this new message and a delimiter. The SecureError and FriendlyError
//message are left untouched.
func (f *Friendly) Extend(msg string, fields ...Fields) *Friendly {
	return &Friendly{
		secureMsg:   f.secureMsg,
		friendlyMsg: f.friendlyMsg,
		msg:         msg + " : " + f.Error(),
		fields:      combineFields(append([]Fields{f.fields}, fields...)...),
	}
}

//WithFriendly returns a copy of err where the friendlyMsg is set to friendlyMsg
func (f *Friendly) WithFriendly(err *Friendly, friendlyMsg string, fields ...Fields) *Friendly {
	if err == nil {
		return NewFriendly(friendlyMsg, fields...)
	}
	return &Friendly{
		secureMsg:   f.secureMsg,
		msg:         f.msg,
		friendlyMsg: friendlyMsg,
		fields:      combineFields(append([]Fields{f.fields}, fields...)...),
	}
}

//WithFriendly returns a copy of err where the Error() is set to msg. See a;so
//Extend, which prepends a new message to the front of the existing message.
func (f *Friendly) WithError(err *Friendly, msg string, fields ...Fields) *Friendly {
	if err == nil {
		return New(msg, fields...)
	}
	return &Friendly{
		secureMsg:   f.secureMsg,
		msg:         msg,
		friendlyMsg: f.friendlyMsg,
		fields:      combineFields(append([]Fields{f.fields}, fields...)...),
	}
}
