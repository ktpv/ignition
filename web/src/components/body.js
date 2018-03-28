import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from 'material-ui/styles'
import Button from 'material-ui/Button'

const styles = theme => ({
  button: {
    margin: theme.spacing.unit
  },
  input: {
    display: 'none'
  }
})

class Body extends React.Component {
  constructor (props) {
    super(props)
    this.state = {
      orgUrl: ''
    }
  }

  componentDidMount () {
    window.fetch('/organization', {
      credentials: 'same-origin'
    }).then(response => response.json())
      .then(response => {
        console.log(response.url)
        this.setState({orgUrl: response.url})
      })
  }

  handleOrgButtonClick = () => {
    window.location = this.state.orgUrl
  }

  render () {
    const { classes } = this.props
    console.log('In render')
    return (
      <div >
        <Button
          variant='raised'
          className={classes.button}
          onClick={this.handleOrgButtonClick}>
            View My Org
        </Button>
      </div>
    )
  }
}

Body.propTypes = {
  classes: PropTypes.object.isRequired
}

export default withStyles(styles)(Body)
