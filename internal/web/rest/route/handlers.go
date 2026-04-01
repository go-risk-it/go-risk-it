package route

import (
	"net/http"

	kernelctx "github.com/go-risk-it/go-risk-it/internal/kernel/ctx"
	restutils "github.com/go-risk-it/go-risk-it/internal/web/rest/utils"
)

// PlainHandler handles requests with standard http types, returning errors.
type PlainHandler func(http.ResponseWriter, *http.Request) error

// CreateHandler returns an authenticated [Route] that decodes a JSON request body
// into Req, passes it with the [kernelctx.UserContext] to perform, and writes the
// result as a 201 JSON response. Auth, decode, and handler errors are translated
// to HTTP responses via [WrapErrors].
func CreateHandler[Req, Resp any](
	pattern string,
	perform func(kernelctx.UserContext, Req) (Resp, error),
) *Route {
	return Authed(pattern, func(writer http.ResponseWriter, request *http.Request) error {
		userCtx, err := ExtractUserContext(request)
		if err != nil {
			return err
		}

		req, err := restutils.DecodeRequest[Req](writer, request)
		if err != nil {
			return err
		}

		resp, err := perform(userCtx, req)
		if err != nil {
			return err
		}

		return restutils.WriteJSON(writer, http.StatusCreated, resp)
	})
}

// QueryHandler returns an authenticated [Route] that extracts the
// [kernelctx.UserContext], calls query, and writes the result as a 200 JSON
// response. Auth and handler errors are translated to HTTP responses via
// [WrapErrors].
func QueryHandler[Resp any](
	pattern string,
	query func(kernelctx.UserContext) (Resp, error),
) *Route {
	return Authed(pattern, func(writer http.ResponseWriter, request *http.Request) error {
		userCtx, err := ExtractUserContext(request)
		if err != nil {
			return err
		}

		resp, err := query(userCtx)
		if err != nil {
			return err
		}

		return restutils.WriteJSON(writer, http.StatusOK, resp)
	})
}

// DomainCommand returns an authenticated [Route] that builds a domain context C
// from the [kernelctx.UserContext] and the {id} path parameter, decodes a JSON
// request body into Req, and calls perform. On success it returns 204 with no
// body. Context-building, decode, and handler errors are translated to HTTP
// responses via [WrapErrors].
func DomainCommand[C, Req any](
	pattern string,
	withID func(kernelctx.UserContext, int64) C,
	perform func(C, Req) error,
) *Route {
	return Domain(pattern, func(r *http.Request) (C, error) {
		return BuildDomainContext(r, withID)
	}, func(
		writer http.ResponseWriter,
		request *http.Request,
		ctx C,
	) error {
		req, err := restutils.DecodeRequest[Req](writer, request)
		if err != nil {
			return err
		}

		if err := perform(ctx, req); err != nil {
			return err
		}

		writer.WriteHeader(http.StatusNoContent)

		return nil
	})
}

// DomainVoid returns an authenticated [Route] that builds a domain context C
// from the [kernelctx.UserContext] and the {id} path parameter, and calls
// perform with no request body. On success it returns 204 with no body.
// Context-building and handler errors are translated to HTTP responses via
// [WrapErrors].
func DomainVoid[C any](
	pattern string,
	withID func(kernelctx.UserContext, int64) C,
	perform func(C) error,
) *Route {
	return Domain(pattern, func(r *http.Request) (C, error) {
		return BuildDomainContext(r, withID)
	}, func(
		writer http.ResponseWriter,
		_ *http.Request,
		ctx C,
	) error {
		if err := perform(ctx); err != nil {
			return err
		}

		writer.WriteHeader(http.StatusNoContent)

		return nil
	})
}
