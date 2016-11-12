var webpack = require('webpack');
var copy = require('copy-webpack-plugin');
var path = require('path');

module.exports = {
    entry: "./app/index.tsx",
    output: {
        filename: "./dist/bundle.js",
    },

    // Enable sourcemaps for debugging webpack's output.
    devtool: "source-map",

    resolve: {
        // Add '.ts' and '.tsx' as resolvable extensions.
        extensions: ["", ".webpack.js", ".web.js", ".ts", ".tsx", ".js"]
    },

    module: {
        loaders: [
            // All files with a '.ts' or '.tsx' extension will be handled by 'ts-loader'.
            { test: /\.tsx?$/, loader: "ts-loader" }
        ],

        preLoaders: [
            // All output '.js' files will have any sourcemaps re-processed by 'source-map-loader'.
            { test: /\.js$/, loader: "source-map-loader" }
        ]
    },

    // When importing a module whose path matches one of the following, just
    // assume a corresponding global variable exists and use that instead.
    // This is important because it allows us to avoid bundling all of our
    // dependencies, which allows browsers to cache those libraries between builds.
    externals: {
        "react": "React",
        "react-dom": "ReactDOM",
        "grecaptcha": undefined,
    },

    plugins: [
      new copy([
        { from: 'index.html', to: 'dist/' },
        { from: 'assets/', to: 'dist/assets/' },
        { from: 'semantic/dist/', to: 'dist/semantic/' },
        { from: 'node_modules/react/dist/react.min.js', to: 'dist/' },
        { from: 'node_modules/react-dom/dist/react-dom.min.js', to: 'dist/' },
      ]),
    ],
};

