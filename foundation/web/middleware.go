package web

// O trabalho do middleware é receber um handler e retornar um handler
// acrescido de código ou sobre o qual algo foi feito
// Passamos a função que queremos que seja executada para dentro de outra
// com a "promessa" que ela será executada no fim
type Middleware func(Handler) Handler

// recebemos um conjunto de funções que serão executadas uma sobre as outras
// de forma que a inserida por último, usará a penultima e assim até chegar
// na primeira da lista
func wrapMiddleware(mw []Middleware, handler Handler) Handler {
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}

	return handler
}
