import React from 'react'
import PropTypes from 'prop-types'
import { withStyles } from 'material-ui/styles'
import Button from 'material-ui/Button'
import Footer from './footer'
import { getOrgUrl } from '../org'

import milkyWay from './../../images/bkgd_milky-way_full.svg'
import deepSpace from './../../images/bkgd_lvl2_deep-space.svg'
import icePlanet from './../../images/bkgd_lvl3_ice-planet.svg'

import rocketMan from './../../images/frgd_rocket-man.svg'
import moonMan from './../../images/frgd_moon-man.svg'
import pewPew from './../../images/frgd_pewpew-man2.svg'

import step1 from './../../images/step-1.svg'
import step2 from './../../images/step-2.svg'
import step3 from './../../images/step-3.svg'
import pivotalLogo from './../../images/pivotal.png'
import arrowIcon from './../../images/icon_arrow.svg'

const makeSpeechBubbleClass = (bgColor, fgColor) => ({
  position: 'relative', // so we can overlap the button

  padding: '24px',
  borderRadius: '15px',
  backgroundColor: bgColor,
  color: fgColor,

  fontSize: '1.75rem',
  height: 'auto',
  width: '40vw',
  display: 'flex',
  flexDirection: 'column',
  alignItems: 'center',

  '&:before': {
    content: '""',
    width: '0px',
    height: '0px',
    position: 'absolute',
    borderLeft: `100px solid ${bgColor}`,
    borderRight: '100px solid transparent',
    borderTop: `25px solid ${bgColor}`,
    borderBottom: '25px solid transparent',
    right: '-175px',
    top: '75px'
  }
})

const speech1Background = '#083c61'
const speech2Background = '#9bd2d2'
const greenButton = '#007D69'

const footerLinks = [
  { text: 'Copyright', url: '' },
  { text: 'Terms', url: '' },
  { text: 'Contact', url: '' }
]

const styles = theme => ({
  body: {
    fontFamily: 'Roboto, Helvetica, Arial, sans-serif',
    fontWeight: 'lighter',
    marginTop: '68px'
  },
  button: {
    margin: theme.spacing.unit
  },
  // for buttons that overlap the bottom of a speech bubble
  speechButton: {
    position: 'absolute',
    bottom: -4 * theme.spacing.unit,
    backgroundColor: greenButton,
    color: 'white',
    height: 7 * theme.spacing.unit,
    fontWeight: 'bold',
    letterSpacing: '3px',
    boxShadow: '-5px 5px 3px rgba(0, 0, 0, 0.29)'
  },
  welcomeSpeech: makeSpeechBubbleClass(speech1Background, 'white'),
  spacesSpeech: makeSpeechBubbleClass(speech2Background, 'black'),

  // CTA 1: Welcome
  ctaWelcome: {
    backgroundImage: `url("${milkyWay}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'center',
    backgroundColor: '#00253e',
    backgroundSize: 'cover',
    height: '700px',
    padding: 6 * theme.spacing.unit,

    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'center'
  },
  rocketMan: {
    backgroundImage: `url("${rocketMan}")`,
    height: '450px',
    width: '450px',
    backgroundRepeat: 'no-repeat'
  },

  // CTA 2: three steps
  ctaSteps: {
    position: 'relative', // for arrow icon absolute positioning

    backgroundImage: `url("${deepSpace}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'center',
    backgroundSize: 'cover',
    backgroundColor: '#00253e',
    color: 'white',
    padding: 6 * theme.spacing.unit,
    height: '700px',
    fontSize: '32px',
    display: 'flex',
    flexDirection: 'row',
    minWidth: '1224px'
  },
  pewPew: {
    backgroundImage: `url("${pewPew}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'left',
    backgroundSize: 'auto',
    height: '450px',
    width: '450px',
    flexShrink: 0
  },
  step: {
    textAlign: 'center',
    borderTop: '9px solid #FFC712',
    marginTop: '64px',
    flexGrow: '1',
    '& p': {
      maxWidth: '258px',
      margin: 'auto'
    }
  },
  stepImage: {
    height: '146px',
    marginTop: '-82px'
  },

  // CTA 3: spaces overview
  ctaSpaces: {
    backgroundImage: `url("${icePlanet}")`,
    backgroundRepeat: 'no-repeat',
    backgroundPosition: 'center',
    backgroundSize: 'cover',
    height: '700px',
    padding: 6 * theme.spacing.unit,
    display: 'flex',
    flexDirection: 'row',
    justifyContent: 'space-around'
  },
  moonMan: {
    backgroundImage: `url("${moonMan}")`,
    height: '450px',
    width: '400px',
    backgroundRepeat: 'no-repeat'
  },
  temporary: {
    alignItems: 'center'
  },
  arrow: {
    position: 'absolute',
    bottom: '10px',
    height: '100px',
    marginLeft: '45%'
  }
})

class Body extends React.Component {
  constructor (props) {
    super(props)
    this.state = {
      orgUrl: ''
    }
  }

  handleOrgButtonClick = async () => {
    // TODO: show spinner
    const url = await getOrgUrl()
    if (url) {
      this.setState({ orgUrl: url })
      window.location = url
    }
  }

  renderWelcomeInfo () {
    const { classes } = this.props

    return (
      <div className={classes.ctaWelcome}>
        <div>
          <div className={classes.welcomeSpeech}>
            {introMessages.map((msg, i) => <p key={i}>{msg}</p>)}
            {this.renderButton('Give Me an Org!', classes.speechButton)}
          </div>
        </div>
        <div className={classes.rocketMan} />
      </div>
    )
  }

  renderGettingStartedSteps () {
    const { classes } = this.props
    return (
      <div className={classes.ctaSteps}>
        <div className={classes.pewPew} />
        <div className={classes.step}>
          <div>
            <img className={classes.stepImage} src={step1} />
          </div>
          <p>
            Get the<br />
            <a href="https://docs.pivotal.io/pivotalcf/latest/cf-cli/">
              Cloud Foundry CLI
            </a>
            <br />
            from Pivotal
          </p>
        </div>
        <div className={classes.step}>
          <div>
            <img className={classes.stepImage} src={step2} />
          </div>
          <p>
            Download the <br />
            <a href="https://github.com/cloudfoundry-samples/spring-music">
              sample app
            </a>
            <br />
            from Github
          </p>
        </div>
        <div className={classes.step}>
          <div>
            <img className={classes.stepImage} src={step3} />
          </div>
          <p>
            Learn to<br />
            <a href="https://docs.pivotal.io/pivotalcf/latest/devguide/deploy-apps/deploy-app.html">
              deploy an app
            </a>
          </p>
        </div>
        <img className={classes.arrow} src={arrowIcon} />
      </div>
    )
  }

  renderSpacesInfo () {
    const { classes } = this.props
    return (
      <div className={classes.ctaSpaces}>
        <div>
          <div className={classes.spacesSpeech}>
            {spaceMessages.map((msg, i) => <p key={i}>{msg}</p>)}
            {this.renderButton(
              `I'm ready. Go to my org!`,
              classes.speechButton
            )}
          </div>
        </div>
        <div className={classes.moonMan} />
      </div>
    )
  }

  renderButton (text, extraClasses) {
    let classes = this.props.classes.button
    if (extraClasses) classes += ' ' + extraClasses
    return (
      <Button
        size="large"
        variant="raised"
        className={classes}
        onClick={this.handleOrgButtonClick}
      >
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
        <Footer links={footerLinks} logoURL={pivotalLogo} />
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
