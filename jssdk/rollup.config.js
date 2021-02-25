import minify from "rollup-plugin-babel-minify";
import nodeResolve from 'rollup-plugin-node-resolve';
import { version } from './package.json'

export default {
  input: 'src/index.js',
  plugins: [
    nodeResolve({ jsnext: true, main: false })
  ],
  output: [{
    name: 'CS',
    file: `../examples/chat/cssdk.js`,
    format: 'umd',
    // exports: 'named'
  }, {
    name: 'CS',
    file: `./dist/cssdk.js`,
    format: 'umd',
    // exports: 'named'
  }, {
    compact: true,
    plugins: [minify({ comments: false })],
    name: 'CS',
    file: `./dist/cssdk.min.js`,
    format: 'umd',
    // exports: 'named'
  }],
}