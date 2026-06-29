// Package arklog provides the small opt-in logger used by parser and example
// code.
//
// Logging is disabled by default per level. Callers can enable individual
// levels or the aggregate All level when they need parser/API diagnostics while
// keeping normal command output quiet.
package arklog
