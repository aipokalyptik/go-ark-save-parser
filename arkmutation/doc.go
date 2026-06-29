// Package arkmutation contains experimental copied-save mutation helpers.
//
// Mutation functions always require an explicit output path and never modify
// the input save in place. The helpers are structurally tested by reopening and
// inspecting copied saves where feasible, but generated saves are
// live-server-unverified and should be treated as experimental until manually
// validated against disposable server data.
package arkmutation
