import React from 'react'
import AppBar from './app-bar'
import Body from './body'
import withRoot from '../withRoot'

const Home = () => (
  <div>
    <AppBar />
    <Body />
  </div>
)

export default withRoot(Home)
