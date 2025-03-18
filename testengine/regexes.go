package testengine

import (
	"regexp"
)

var timeRegexes = []*regexp.Regexp{
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`),                                            // ISO 8601 / RFC 3339, UTC
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[\+\-]\d{2}:\d{2}$`),                            // ISO 8601 / RFC 3339, with offset
	regexp.MustCompile(`^\d{2}:\d{2}$`),                                                                     // 24-Hour Time, HH:MM
	regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`),                                                               // 24-Hour Time, HH:MM:SS
	regexp.MustCompile(`^(1[0-2]|0?[1-9]):[0-5]\d (AM|PM)$`),                                                // 12-Hour Time, HH:MM AM/PM
	regexp.MustCompile(`^(1[0-2]|0?[1-9]):[0-5]\d:[0-5]\d (AM|PM)$`),                                        // 12-Hour Time, HH:MM:SS AM/PM
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),                                                               // Complete Date, YYYY-MM-DD
	regexp.MustCompile(`(?i)^(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec) \d{2}, \d{4}$`),              // Month Day, Year, MMM DD, YYYY
	regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}$`),                                                         // 2024-02-14T12
	regexp.MustCompile(`(?i)^\d{1,2}-(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-(?:\d{2}|\d{4})$`), // 25-May-2007
	regexp.MustCompile(`(?i)^(?:Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-(?:\d{2}|\d{4})$`),         // May-25-2007
}
