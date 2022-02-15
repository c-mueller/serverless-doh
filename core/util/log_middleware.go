package util

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"strings"
	"time"
)

func LogMiddleware(moduleName string, rules []RegexRule, log *logrus.Entry, verbose bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		method := c.Request.Method

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		clientIP := c.ClientIP()
		statusCode := c.Writer.Status()

		comment := c.Errors.ByType(gin.ErrorTypePrivate).String()

		if raw != "" {
			path = path + "?" + raw
		}

		lat := float64(latency.Nanoseconds()/1000) / 1000
		message := fmt.Sprintf("[%s]: %s %s - %d: %.3f MS %s", clientIP, method, path, statusCode, lat, comment)

		messageb := []byte(message)

		for _, v := range rules {
			messageb = v.apply(messageb, true)
		}

		l := log.WithFields(logrus.Fields{
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"status_code": statusCode,
			"latency":     lat,
			"comment":     comment,
		})

		if verbose {
			reqH := make(map[string]string)
			for k, v := range c.Request.Header {
				sb := strings.Builder{}
				for i, v := range v {
					if i != 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(v)
				}
				reqH[k] = sb.String()
			}
			resH := make(map[string]string)
			for k, v := range c.Writer.Header() {
				sb := strings.Builder{}
				for i, v := range v {
					if i != 0 {
						sb.WriteString(", ")
					}
					sb.WriteString(v)
				}
				resH[k] = sb.String()
			}
			l = l.WithFields(logrus.Fields{
				"request_headers":  reqH,
				"response_headers": resH,
				"full_path":        path,
			})
		}

		l.Info(string(messageb))
	}
}
