import sys

class ErrorCode:
    def __init__(self, name, code, extra=None):
        self.name = name
        self.code = code
        self.extra = extra
        self.categories = []

class ErrorClass:
    def __init__(self, name, codes):
        self.name = name
        self.codes = codes


def parse_error_definitions_from_file(errors_filename):
    errors_file = open(errors_filename, 'r')
    errors_code = compile(errors_file.read(), errors_filename, 'exec')
    error_codes = []
    error_classes = []
    eval(errors_code,
            dict(error_code=lambda *args, **kw: error_codes.append(ErrorCode(*args, **kw)),
                 error_class=lambda *args: error_classes.append(ErrorClass(*args))))
    error_codes.sort(key=lambda x: x.code)

    return error_codes, error_classes

errorCodes, errorClasses = parse_error_definitions_from_file(sys.argv[1])


# tempalte out go classes!
print 'package mongoerror'
print
print 'import "go.mongodb.org/mongo-driver/bson"'
print
print 'type ErrorCode int'
print
print 'const ('
for c in errorCodes:
    print '	%s ErrorCode = %d' % (c.name, c.code)
print ')'
print
print 'func (c ErrorCode) String() string {'
print '	switch c {'
for c in errorCodes:
    print '	case %d:' % (c.code)
    print '		return "%s"' % (c.name)
print '	default:'
print '		panic("Unknown")'
print '	}'
print '}'
print
print 'func (c ErrorCode) ErrMessage(msg string) bson.D {'
print '	r := bson.D{{"ok", 0}, {"errmsg", msg}, {"code", int(c)}, {"codeName", c.String()}}'
print '	return r'
print '}'
for c in errorClasses:
    print
    print 'func Is%s(e ErrorCode) bool {' % (c.name)
    print '	switch e {'
    for e in c.codes:
    	print '	case %s:' % (e)
    	print '		return true'
    print '	}'
    print '	return false'
    print '}'
