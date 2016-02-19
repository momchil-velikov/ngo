
builtins = ["BuiltinAppend", "BuiltinCap", "BuiltinClose", "BuiltinComplex",
     "BuiltinCopy", "BuiltinDelete", "BuiltinImag", "BuiltinLen", "BuiltinMake",
     "BuiltinNew", "BuiltinPanic", "BuiltinPrint", "BuiltinPrintln", "BuiltinReal",
     "BuiltinRecover"]


#for fn in builtins:
#    print "func (*infer) infer%s(x*ast.Call) (ast.Expr, error) {\nreturn x, nil\n}\n" % fn


for fn in builtins:
    print "case ast.Builtin%s:\nreturn ctx.infer%s(x)" %(fn, fn)
