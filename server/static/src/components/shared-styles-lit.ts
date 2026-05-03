import { css } from 'lit';

export const sharedStyles = css`
  /* 640px is the size of responsiveWidth of app-drawer, 300px is default drawer width*/
  @media (min-width:940px) {
    .card {
      margin: 24px;
      padding: 16px;
      color: var(--md-sys-color-on-surface, #1C1810);
      border-radius: 12px;
      background: linear-gradient(180deg, var(--md-sys-color-surface-container-low, #F5F0E8) 0%, var(--md-sys-color-surface-container, #F0EBE3) 100%);
      box-shadow: var(--md-sys-elevation-1, 0 1px 3px 0 rgba(60,40,20,.10), 0 1px 2px 0 rgba(60,40,20,.06)),
                  inset 0 1px 0 rgba(255, 255, 255, 0.5);
    }
  }

  @media (max-width:940px) {
    .card {
      padding: 16px;
      color: var(--md-sys-color-on-surface, #1C1810);
      background-color: var(--md-sys-color-surface-container-low, #F5F0E8);
      border-bottom: 1px solid var(--md-sys-color-outline-variant, #CCC4B8);
    }
  }

  .circle {
    display: inline-block;
    width: 64px;
    height: 64px;
    text-align: center;
    color: var(--md-sys-color-on-surface-variant, #4A4539);
    border-radius: 50%;
    background: var(--md-sys-color-surface-container-highest, #E0D9CE);
    font-size: 30px;
    line-height: 64px;
  }

  h1 {
    margin: 16px 0;
    color: var(--md-sys-color-on-surface, #1C1810);
    font-size: 22px;
  }
`;