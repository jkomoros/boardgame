import './boardgame-track.js';
import './boardgame-mat.js';

const track = document.createElement('boardgame-track');
track.label = 'Climate';
track.steps = [{ key: 'warm', label: 'Warm', color: '#fca' }];
track.value = 'warm';
// @ts-expect-error A track key is a stable string or number, not an object.
track.steps = [{ key: {}, label: 'Warm' }];
const mat = document.createElement('boardgame-mat');
mat.label = 'Species';
mat.art = './portrait.jpg';
mat.headingLevel = 3;
// @ts-expect-error A mat heading level is numeric.
mat.headingLevel = '3';
