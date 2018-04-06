import React from 'react'
import PropTypes from 'prop-types'
import Typography from 'material-ui/Typography';
import { withStyles } from 'material-ui/styles'
import Button from 'material-ui/Button'

import milkyWay from './../../images/bkgd_milky-way_full.svg'
import deepSpace from './../../images/bkgd_lvl2_deep-space.svg'
import icePlanet from './../../images/bkgd_lvl3_ice-planet.svg'

import rocketMan from './../../images/frgd_rocket-man.svg'
import moonMan from './../../images/frgd_moon-man.svg'
import pewPew from './../../images/frgd_pewpew-man2.svg'

import step1 from './../../images/step_1.svg'
import step2 from './../../images/step_2.svg'
import step3 from './../../images/step_3.svg'

const bubbleBackground1 = '#083c61'

const styles = theme => ({
  body: {
    fontFamily: 'Roboto, Helvetica, Arial, sans-serif',
    fontWeight: 'lighter'
  },
  button: {
    margin: theme.spacing.unit
  },
  // for buttons that overlap the bottom of a speech bubble
  speechButton: {
    position: 'absolute',
    bottom: -3 * theme.spacing.unit
  },
  speechBubble: {
    position: 'relative', // so we can overlap the button

    padding: '24px',
    borderRadius: '15px',
    '&:before': {
      content: '""',
      width: '0px',
      height: '0px',
      position: 'absolute',
      borderLeft:'100px solid #083c61',
      borderRight:'100px solid transparent',
      borderTop:'25px solid #083c61',
      borderBottom:'25px solid transparent',
      right: '-175px',
      top: '75px'
    }
  },

  // CTA 1: Welcome
  ctaWelcome: {
    backgroundImage: `url("${milkyWay}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'center',
    backgroundColor: '#00253e',
    backgroundSize: '100% auto',
    height: '700px',
    padding: 6 * theme.spacing.unit,

    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'center'
  },
  welcomeSpeech: {
    color: 'white',
    fontSize: '1.75rem',
    height: 'auto',
    width: '40vw',
    borderRadius: '15px',
    backgroundColor: '#083c61'
  },
  rocketMan: {
    backgroundImage: `url("${rocketMan}")`,
    //height: '375px',
    width: '40vw',
    backgroundRepeat: 'no-repeat'
  },

  // CTA 2: three steps
  ctaSteps: {
    backgroundImage: `url("${deepSpace}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'center',
    backgroundSize: '100% 100%',
    backgroundColor: '#00253e',
    color: 'white',
    padding: 6 * theme.spacing.unit,
    height: '700px',
    fontSize: '32px'
  },
  pewPew: {
    backgroundImage: `url("${pewPew}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'left',
    backgroundSize: 'auto',
    height: '450px',

    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-evenly',
    alignItems: 'flex-start',

    paddingLeft: '250px',
    paddingRight: '75px',
    paddingBottom: '50px'
  },
  step: {
    textAlign: 'center',
    width: '225px'
  },
  stepImage: {
    height: '146px',
    //width: '100%'
  },

  // CTA 3: spaces overview
  ctaSpaces: {
    backgroundImage: `url("${icePlanet}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'center',
    backgroundSize: '100% auto',
    height: '700px',
    padding: 6 * theme.spacing.unit,
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-around'
  },
  spacesSpeech: {
    color: 'black',
    fontSize: '2rem',
    height: '375px',
    width: '40vw',
    backgroundColor: '#9ed4d4',
  },
  moonMan: {
    backgroundImage: `url("${moonMan}")`,
    height: '375px',
    width: '400px',
    backgroundRepeat: 'no-repeat'
  }
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
      return response.json()
    }).then(response => {
      if (!response) {
        return
      }
      this.setState({orgUrl: response.url})
      window.location = response.url
    })
  }

  renderWelcomeInfo () {
    const { classes } = this.props

    return (
      <div className={classes.ctaWelcome}>
        <div>
          <div className={classes.welcomeSpeech + ' ' + classes.speechBubble}>
            {
              introMessages.map(msg => <p>{msg}</p>)
            }
            {
              false && this.renderButton('Give Me an Org!', classes.speechButton)
            }
          </div>
        </div>
        <div className={classes.rocketMan}></div>
      </div>
    )
  }

  renderGettingStartedSteps () {
    const { classes } = this.props
    return (
      <div className={classes.ctaSteps}>
        <div className={classes.pewPew}>
          <div className={classes.step}>
            <div><img className={classes.stepImage} src={step1} /></div>
            Get the <a href='https://docs.pivotal.io/pivotalcf/latest/cf-cli/'>Cloud Foundry CLI from Pivotal.</a>
          </div>
          <div className={classes.step}>
            <div><img className={classes.stepImage} src={step2} /></div>
            Download the <a href='https://github.com/cloudfoundry-samples/spring-music'>sample app from Github.</a>
          </div>
          <div className={classes.step}>
            <div><img className={classes.stepImage} src={step3} /></div>
            Learn to <a href='https://docs.pivotal.io/pivotalcf/latest/devguide/deploy-apps/deploy-app.html'>deploy an app.</a>
          </div>
        </div>
      </div>
    )
  }

  renderSpacesInfo () {
    const { classes } = this.props
    return (
      <div className={classes.ctaSpaces}>
        <div className={classes.spacesSpeech + ' ' + classes.speechBubble}>
          {
            spaceMessages.map(msg => <p>{msg}</p>)
          }
        </div>
        <div className={classes.moonMan}></div>
      </div>
    )
  }

  renderButton (text, extraClasses) {
    let classes = this.props.classes.button
    if (extraClasses) classes += ' ' + extraClasses
    return (
      <Button
        variant='raised'
        className={classes}
        onClick={this.handleOrgButtonClick}>
          {text}
      </Button>
    )
  }

  render () {
    const { classes } = this.props
    return (
      <div className={classes.body}>
        {this.renderWelcomeInfo()}
        {this.renderGettingStartedSteps()}
        {this.renderSpacesInfo()}
        {this.renderButton('View My Org')}
      </div>
    )
  }
}

Body.propTypes = {
  classes: PropTypes.object.isRequired
}

const introMessages = [
  // TODO: replace 'Pivotal' with <CompanyName> from component state..
  'Pivotal is giving you a free playground to push (deploy) apps and experiment.  PCF uses orgs to organize things.',
  'Orgs contain spaces, and each space can host apps.  You will get your very own org and can create as many spaces as you like.'
]

const spaceMessages = [
  // TODO: replace 'development' with <SpaceName> from component state...
  'Spaces can act like environments, and your first space is called "development".',
  'Once apps are pushed to a space, you can bind them to services like MySQL and NewRelic by visiting the "Marketplace" link in PCF.'
]

export default withStyles(styles)(Body)
