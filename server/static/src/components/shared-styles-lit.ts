import { css } from 'lit';

export const sharedStyles = css`
  /* 640px is the size of responsiveWidth of app-drawer, 300px is default drawer width*/
  @media (min-width:940px) {
    .card {
      margin: 24px;
      padding: 16px;
      color: #757575;
      border-radius: 5px;
      background-color: #fff;
      box-shadow: 0 2px 2px 0 rgba(0, 0, 0, 0.14), 0 1px 5px 0 rgba(0, 0, 0, 0.12), 0 3px 1px -2px rgba(0, 0, 0, 0.2);
    }
  }

  @media (max-width:940px) {
    .card {
      padding: 16px;
      color: #757575;
      background-color: #fff;
      border-bottom:1px solid #CCCCCC;
    }
  }

  .circle {
    display: inline-block;
    width: 64px;
    height: 64px;
    text-align: center;
    color: #555;
    border-radius: 50%;
    background: #ddd;
    font-size: 30px;
    line-height: 64px;
  }

  h1 {
    margin: 16px 0;
    color: #212121;
    font-size: 22px;
  }
`;