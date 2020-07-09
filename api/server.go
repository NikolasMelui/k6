/*
 *
 * k6 - a next-generation load testing tool
 * Copyright (C) 2016 Load Impact
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package api

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/urfave/negroni"

	"github.com/loadimpact/k6/api/common"
	v1 "github.com/loadimpact/k6/api/v1"
	"github.com/loadimpact/k6/core"
)

func NewHandler(logger logrus.FieldLogger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/v1/", v1.NewHandler())
	mux.Handle("/ping", HandlePing(logger))
	mux.Handle("/", HandlePing(logger))
	return mux
}

func ListenAndServe(addr string, engine *core.Engine, logger logrus.FieldLogger) error {
	mux := NewHandler(logger)

	n := negroni.New()
	n.Use(negroni.NewRecovery())
	n.UseFunc(WithEngine(engine))
	n.UseFunc(NewLogger(logger))
	n.UseHandler(mux)

	return http.ListenAndServe(addr, n)
}

// NewLogger returns the middleware which logs response status for request.
func NewLogger(l logrus.FieldLogger) negroni.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r)

		res := rw.(negroni.ResponseWriter)
		l.WithField("status", res.Status()).Debugf("%s %s", r.Method, r.URL.Path)
	}
}

func WithEngine(engine *core.Engine) negroni.HandlerFunc {
	return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		r = r.WithContext(common.WithEngine(r.Context(), engine))
		next(rw, r)
	})
}

func HandlePing(logger logrus.FieldLogger) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Content-Type", "text/plain; charset=utf-8")
		if _, err := fmt.Fprint(rw, "ok"); err != nil {
			logger.WithError(err).Error("Error while printing ok")
		}
	})
}
