var path = require('path')

var config = {
  mode: 'development',
  devtool: 'source-map',
  entry: {
    index: path.resolve(__dirname, 'src/index.jsx')
  },
  output: {
    path: path.resolve(__dirname, '../assets/script'),
    filename: '[name].js'
  },
  module: {
    rules: [
      {
        test: /\.jsx?$/,
        loader: 'babel-loader',
        exclude: /(node_modules)/,
        query: {
          presets: ['@babel/react'],
          cacheDirectory: true
        }
      }
    ]
  },
  resolve: {
    modules: [
      path.resolve(path.resolve(__dirname, 'src')),
      path.resolve('./node_modules')
    ]
  }
}

module.exports = config