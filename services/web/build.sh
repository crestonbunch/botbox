#!/bin/bash

yarn install

cd semantic
gulp build-css
gulp build-assets
cd ../

echo "export class Settings {
    public static RECAPTCHA_KEY: string = \"$BOTBOX_RECAPTCHA_SITEKEY\";
    public static BOTBOX_GITHUB_ID: string = \"$BOTBOX_GITHUB_ID\";
    public static API_BASE_URL: string = \"$BOTBOX_DOMAIN_NAME/api/\";
}" > app/settings.ts

webpack --config=webpack.prod.config.js

docker build -t botbox-web:$BOTBOX_VERSION .
