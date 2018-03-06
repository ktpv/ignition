const HtmlWebpackPlugin = require('html-webpack-plugin')
const path = require('path')

module.exports = {
  entry: [
    './src/index.js'
  ],
  devtool: 'source-map',
  module: {
    rules: [
      {
        test: /\.(js|jsx)$/,
        exclude: /node_modules/,
        use: ['babel-loader']
      }
    ]
  },
  resolve: {
    extensions: ['*', '.js', '.jsx']
  },
  output: {
    path: path.join(__dirname, '/dist'),
    publicPath: '/',
    filename: 'assets/bundle.js'
  },
  plugins: [
    new HtmlWebpackPlugin({
      title: 'Pivotal Ignition',
      template: 'src/index.html',
      hash: true,
      filename: 'index.html'
    })
  ]
}
