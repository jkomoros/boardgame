/**
 * companion-avatar-catalog defines the placeholder avatar primaries +
 * name vocabulary for the Table+Hand join flow's anonymous avatar
 * picker (spec §10). Modeled on word-bloom's publisher-avatar 4-tuple
 * composite (primary-decoration-corner-tint) but VASTLY simplified for
 * V1 — just emoji primaries + ASCII adjective+noun names.
 *
 * Swap in a real SVG catalog (12 primaries × 12 decorations × 4 corners
 * × 8 tints = 4,608 composites) when art direction is decided. The
 * AvatarSlug returned by randomAvatarSlug() and rendered by glyph-for-
 * slug must keep its public interface (string slug, fits in
 * seatPresentation.AvatarSlug's 256 chars) when the catalog is swapped.
 */

export const PRIMARIES = [
  '🦊', '🐻', '🦁', '🐯', '🐸', '🐙', '🦄', '🐳', '🦉', '🐧', '🐲', '🦋',
  '🐺', '🦅', '🦈', '🐬', '🦎', '🐢', '🦩', '🐝', '🐞', '🦇', '🐠', '🦜',
];

export const ADJECTIVES = [
  'Brave', 'Clever', 'Sunny', 'Wild', 'Bright', 'Mighty', 'Calm', 'Bold',
  'Quick', 'Quiet', 'Shy', 'Lucky', 'Cosmic', 'Daring', 'Eager', 'Fierce',
  'Jolly', 'Noble', 'Swift', 'Witty', 'Grand', 'Keen', 'Plucky', 'Wry',
  'Lively', 'Merry', 'Nimble', 'Royal', 'Sly', 'Zesty',
];

export const NOUNS = [
  'Fox', 'Bear', 'Lion', 'Tiger', 'Frog', 'Octopus', 'Unicorn', 'Whale',
  'Owl', 'Penguin', 'Dragon', 'Butterfly', 'Phoenix', 'Wolf', 'Otter', 'Hawk',
  'Raven', 'Shark', 'Dolphin', 'Lizard', 'Turtle', 'Falcon', 'Badger', 'Heron',
  'Crane', 'Panther', 'Cobra', 'Sparrow', 'Beetle', 'Parrot',
];

function randomFromArray<T>(arr: T[]): T {
  return arr[Math.floor(Math.random() * arr.length)];
}

/** Return a fresh random composite avatar slug. V1: just a primary glyph. */
export function randomAvatarSlug(): string {
  return randomFromArray(PRIMARIES);
}

/** Render a slug as a display glyph. V1: identity (the slug IS the glyph). */
export function glyphForSlug(slug: string): string {
  return slug;
}

/** Random adjective+noun display name. */
export function randomDisplayName(): string {
  return randomFromArray(ADJECTIVES) + randomFromArray(NOUNS);
}
