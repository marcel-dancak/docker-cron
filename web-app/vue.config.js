const CompressionPlugin = require('compression-webpack-plugin')

module.exports = {
  lintOnSave: 'warning',
  transpileDependencies: [
    'vuetify'
  ],
  publicPath: process.env.NODE_ENV === 'production' ? '/ui/' : '/',
  assetsDir: 'static',
  configureWebpack: {
    plugins: [new CompressionPlugin()]
  },
  chainWebpack: config => {
    const svgRule = config.module.rule('svg')
    svgRule.uses.clear()

    config.module
      .rule('svg')
      .oneOf('sprite')
        .test(/icons\/.*\.svg$/)
        .use('babel')
          .loader('babel-loader')
          .end()
        .use('svg-sprite')
          .loader('svg-sprite-loader')
          .end()
        .use('svgo')
          .loader('svgo-loader')
          .end()
        .end()

      .oneOf('inline-svg')
        .resourceQuery(/inline/)
        .use('babel')
          .loader('babel-loader')
          .end()
        .use('vue-svg-loader')
          .loader('vue-svg-loader')
          .end()
        .end()

      .oneOf('other')
        .use('file-loader')
          .loader('file-loader')
          .options({
            name: 'static/img/[name].[hash:8].[ext]'
          })
          .end()
        .end()
  },
  devServer: {
    compress: true,
    proxy: {
      '^/api': {
        target: 'http://localhost:8090'
      },
      '^/ws': {
        target: 'ws://localhost:8090',
        secure: false,
        ws: true
      }
    }
  }
}
