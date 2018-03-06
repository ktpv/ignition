import React from 'react'
import {shallow} from 'enzyme'
import Hello from './../../src/components/hello'

test('hello displays a greeting', () => {
  // Render a checkbox with label in the document
  const hello = shallow(<Hello />)
  expect(hello.text()).toEqual('Hello, world!')
})
