// Copyright (C) 2017 JT Olds
// See LICENSE for copying information.

package sheepda

// ChurchNumeral returns a Church-encoded representation of the number val.
func ChurchNumeral(val uint) Value {
	var base Expr = &VariableExpr{Name: "x"}
	for i := uint(0); i < val; i++ {
		base = &ApplicationExpr{
			Func: &VariableExpr{Name: "f"},
			Arg:  base}
	}
	return NewClosure(NewScope(), &LambdaExpr{
		Arg: "f",
		Body: &LambdaExpr{
			Arg:  "x",
			Body: base}})
}

// ChurchPair returns a Church-encoded pair of the two values first and second.
func ChurchPair(first, second Value) Value {
	return NewClosure(NewScope().
		Set("first", first).
		Set("second", second),
		&LambdaExpr{
			Arg: "p",
			Body: &ApplicationExpr{
				Func: &ApplicationExpr{
					Func: &VariableExpr{Name: "p"},
					Arg:  &VariableExpr{Name: "first"}},
				Arg: &VariableExpr{Name: "second"}}})
}

var (
	churchTrue = NewClosure(NewScope(), &LambdaExpr{
		Arg: "t",
		Body: &LambdaExpr{
			Arg:  "f",
			Body: &VariableExpr{Name: "t"}}})
	churchFalse = NewClosure(NewScope(), &LambdaExpr{
		Arg: "t",
		Body: &LambdaExpr{
			Arg:  "f",
			Body: &VariableExpr{Name: "f"}}})
)

// ChurchBool returns a Church-encoded boolean representation of val.
func ChurchBool(val bool) Value {
	if val {
		return churchTrue
	}
	return churchFalse
}
