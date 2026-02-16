/**
 * Global type declarations for variables defined in index.html
 */

interface Config {
  firebase: any;
  offline_dev_mode?: boolean;
}

declare const CONFIG: Config;
declare const API_HOST: string;
