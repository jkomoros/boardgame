/*
Package legal aliases boardgame's Legal*-prefixed value types (Outcome,
BindingValue, Message, Verdict, Cost, Facet, PropPath, Read, Spec) so game
authors write legal.Spec rather than boardgame.LegalSpec, and provides
Verdict constructors (PassVerdict, FailT, UnknownVerdict) and BindingValue
helpers (String, Int, Bool). Core owns the underlying types and, in later
tasks, the evaluation engine; this package will grow into the predicate
catalog and registry, peer to constraints. See
docs/superpowers/specs/2026-07-10-declarative-legality-design.md for the
full design.
*/
package legal
