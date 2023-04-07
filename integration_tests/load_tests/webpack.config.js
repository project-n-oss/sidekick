const path = require("path");
const webpack = require("webpack"); // only add this if you don't have yet
require("dotenv").config({ path: "./.env" });

module.exports = {
  mode: "production",
  entry: {
    get: "./src/get.test.js",
  },
  output: {
    path: path.resolve(__dirname, "dist"), // eslint-disable-line
    libraryTarget: "commonjs",
    filename: "[name].bundle.js",
  },
  module: {
    rules: [{ test: /\.js$/, use: "babel-loader" }],
  },
  target: "web",
  externals: /k6(\/.*)?/,
  plugins: [
    new webpack.DefinePlugin({
      "process.env": JSON.stringify(process.env),
    }),
  ],
};
