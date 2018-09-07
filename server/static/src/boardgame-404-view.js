/**
@license
Copyright (c) 2016 The Polymer Project Authors. All rights reserved.
This code may only be used under the BSD style license found at http://polymer.github.io/LICENSE.txt
The complete set of authors may be found at http://polymer.github.io/AUTHORS.txt
The complete set of contributors may be found at http://polymer.github.io/CONTRIBUTORS.txt
Code distributed by Google as part of the polymer project is also
subject to an additional IP rights grant found at http://polymer.github.io/PATENTS.txt
*/
import { Element } from '@polymer/polymer/polymer-element.js';

import { html } from '@polymer/polymer/lib/utils/html-tag.js';

class Boardgame404View extends Element {
  static get template() {
    return html`
    <style>
      :host {
        display: block;

        padding: 10px 20px;
      }
    </style>

    <!-- 
      If deploying in a folder replace href="/" with the full path to your site.
      Such as: href=="http://polymerelements.github.io/polymer-starter-kit"
    -->
    Oops you hit a 404. <a href="/">Head back to home.</a>
`;
  }

  static get is() {
    return "boardgame-404-view"
  }
}

customElements.define(Boardgame404View.is, Boardgame404View);
