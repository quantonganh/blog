package http

import (
	"context"
	"regexp"
)

func (s *Server) ProcessActivityStream(ctx context.Context, token string) error {
	events, err := s.EventService.Consume(ctx, "page-views")
	if err != nil {
		return err
	}

	if err := s.StatService.ImportIP2LocationDB(token); err != nil {
		return err
	}

	for e := range events {
		country, err := s.StatService.GetCountryFromIP(e.IP)
		if err != nil {
			e.Country = "Unknown"
		}
		e.Country = country
		e.Browser = getBrowser(e.UserAgent)
		e.OS = getOS(e.UserAgent)
		if err := s.StatService.Insert(e); err != nil {
			s.logger.Error().Err(err).Msg("failed to insert event")
		}
	}

	return nil
}

func getBrowser(ua string) string {
	var browserRegexp = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(firefox|fxios)\/\S+`),
		regexp.MustCompile(`(?i)(chrome|chromium|crios)\/\S+`),
		regexp.MustCompile(`(?i)version\/\S+ (safari)\/\S+`),
		regexp.MustCompile(`(?i)(opera|opr)\/\S+`),
	}
	for _, re := range browserRegexp {
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return matches[1] // Return the first matching group which should be the browser name
		}
	}
	return "Unknown"
}

func getOS(ua string) string {
	var osRegexp = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(Windows NT)`),
		regexp.MustCompile(`(?i)(Mac OS X)`),
		regexp.MustCompile(`(?i)(Linux)`),
		regexp.MustCompile(`(?i)(Android)`),
		regexp.MustCompile(`(?i)(iOS)`),
	}
	for _, re := range osRegexp {
		matches := re.FindStringSubmatch(ua)
		if len(matches) > 1 {
			return matches[1] // Return the first matching group which should be the OS
		}
	}
	return "Unknown"
}
