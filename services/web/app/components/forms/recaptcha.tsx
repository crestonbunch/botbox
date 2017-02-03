import * as React from "react";
import 'grecaptcha'
import { Icon } from "semantic-ui-react";

const RECAPTCHA_ID = "recaptcha";

export interface RecaptchaProps {
  sitekey: string,
  onChange: (val: string) => void,
}

export interface RecaptchaState {
  ready?: boolean,
  readyCheck?: any,
}

export class Recaptcha extends React.Component<RecaptchaProps, RecaptchaState> {

  constructor(props: RecaptchaProps) {
    super(props);
    this.state = {
      ready: false,
      readyCheck: setInterval(this.checkReady.bind(this), 200),
    }
  }

  private renderRecaptcha() {
    if (this.state.ready) {
      this.forceUpdate();
      grecaptcha.render(
        RECAPTCHA_ID,
        {
          sitekey: this.props.sitekey,
          callback: this.props.onChange,
        }
      );
    }
  }

  private checkReady() {
    if (grecaptcha != undefined) {
      this.setState({
        ready: true,
      });
      this.renderRecaptcha();
      clearInterval(this.state.readyCheck);
    }
  }

  reset() {
    if (grecaptcha != undefined) {
      grecaptcha.reset();
    }
  }

  render() {
    if (!this.state.ready) {
      return <Icon size="large" name="notched circle" loading />
    }
    return (
        <div id={RECAPTCHA_ID} data-sitekey={this.props.sitekey}></div>
    );
  }
}

