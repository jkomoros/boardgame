import { css } from 'lit';

export const sharedStyles = css`
  /* 640px is the size of responsiveWidth of app-drawer, 300px is default drawer width*/
  @media (min-width:940px) {
    .card {
      margin: 24px;
      padding: 16px;
      color: var(--md-sys-color-on-surface, #1c1b1f);
      border-radius: 12px;
      background-color: var(--md-sys-color-surface-container-low, #f7f2fa);
      box-shadow: var(--md-sys-elevation-1, 0 1px 3px 1px rgba(0,0,0,.15), 0 1px 2px rgba(0,0,0,.3));
    }
  }

  @media (max-width:940px) {
    .card {
      padding: 16px;
      color: var(--md-sys-color-on-surface, #1c1b1f);
      background-color: var(--md-sys-color-surface-container-low, #f7f2fa);
      border-bottom: 1px solid var(--md-sys-color-outline-variant, #cac4d0);
    }
  }

  .circle {
    display: inline-block;
    width: 64px;
    height: 64px;
    text-align: center;
    color: var(--md-sys-color-on-surface-variant, #555);
    border-radius: 50%;
    background: #ddd;
    font-size: 30px;
    line-height: 64px;
  }

  h1 {
    margin: 16px 0;
    color: var(--md-sys-color-on-surface, #212121);
    font-size: 22px;
  }
`;