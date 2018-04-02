import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from 'material-ui/styles'
import Button from 'material-ui/Button'

const styles = theme => ({
  button: {
    margin: theme.spacing.unit
  },
})

class Body extends React.Component {
  constructor (props) {
    super(props)
    this.state = {
      orgUrl: ''
    }
  }

  handleOrgButtonClick = () => {
    // Spinner
    window.fetch('/organization', {
      credentials: 'same-origin'
    }).then(response => {
        if (!response.ok) {
          return
        } 
        const url = response.json().url
        this.setState({orgUrl: url})
        window.location = url
      })
  }

  render () {
    const { classes } = this.props
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
