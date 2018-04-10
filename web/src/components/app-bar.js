import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from 'material-ui/styles'
import AppBar from 'material-ui/AppBar'
import Toolbar from 'material-ui/Toolbar'
import Typography from 'material-ui/Typography'
import IconButton from 'material-ui/IconButton'
import AccountCircle from 'material-ui-icons/AccountCircle'
import Menu, { MenuItem } from 'material-ui/Menu'

import ignitionLogo from './../../images/ignition.svg'

const styles = {
  root: {
    display: 'flex',
    position: 'sticky',
    top: 0,
    left: 'auto',
    right: 0,
    zIndex: 999
  },
  logoContainer: {
    background: '#F2F0F1',
    padding: 0
  },
  logo: {
    height: '64px',
    padding: '8px 24px'
  },
  userContainer: {
    display: 'flex',
    flexGrow: 1,
    alignItems: 'center'
  },
  name: {
    flexGrow: 1
  },
  icon: {
    flexShrink: 1
  },
  menuButton: {
    marginLeft: -12,
    marginRight: 20
  }
}

class MenuAppBar extends React.Component {
  constructor (props) {
    super(props)
    const { profile } = props
    this.state = {
      auth: true,
      anchorEl: null,
      profile: profile
    }
  }

  handleMenu = event => {
    this.setState({ anchorEl: event.currentTarget })
  }

  handleClose = () => {
    this.setState({ anchorEl: null })
  }

  handleLogout = (e, location = window.location) => {
    this.setState({ anchorEl: null, profile: null })
    if (this.props && this.props.testing) {
      return
    }
    location.replace('/logout')
  }

  componentDidMount () {
    if (this.props && this.props.testing) {
      return
    }
    window
      .fetch('/profile', {
        credentials: 'same-origin'
      })
      .then(response => {
        if (!response.ok) {
          if (response.status === 401) {
            window.location.replace('/login')
            return
          }
          window.location.replace('/' + response.status)
          return
        }
        return response.json()
      })
      .then(profile => this.setState({ profile }))
  }

  render () {
    const { classes } = this.props
    const { anchorEl, profile } = this.state
    const open = Boolean(anchorEl)
    let name = ''
    if (profile && profile.Name) {
      name = profile.Name
    }

    return (
      <div className={classes.root}>
        <AppBar color="white">
          <Toolbar disableGutters={true}>
            <div className={classes.logoContainer}>
              <img className={classes.logo} src={ignitionLogo} />
            </div>
            {profile && (
              <div className={classes.userContainer}>
                <Typography
                  variant="subheading"
                  color="primary"
                  align="right"
                  className={classes.name}
                >
                  {`Welcome, ${name}`}
                </Typography>
                <IconButton
                  aria-owns={open ? 'menu-appbar' : null}
                  aria-haspopup="true"
                  onClick={this.handleMenu}
                  color="primary"
                  className={classes.icon}
                >
                  <AccountCircle />
                </IconButton>
                <Menu
                  id="menu-appbar"
                  anchorEl={anchorEl}
                  anchorOrigin={{
                    vertical: 'top',
                    horizontal: 'right'
                  }}
                  transformOrigin={{
                    vertical: 'top',
                    horizontal: 'right'
                  }}
                  open={open}
                  onClose={this.handleClose}
                >
                  <MenuItem onClick={this.handleLogout}>Logout</MenuItem>
                </Menu>
              </div>
            )}
          </Toolbar>
        </AppBar>
      </div>
    )
  }
}

MenuAppBar.propTypes = {
  classes: PropTypes.object.isRequired,
  testing: PropTypes.bool,
  profile: PropTypes.object
}

MenuAppBar.propTypes = {
  classes: PropTypes.object.isRequired
}

export default withStyles(styles)(MenuAppBar)
