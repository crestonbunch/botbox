import * as React from "react";
import 'grecaptcha'
import { Icon } from "semantic-ui-react";

export interface RecaptchaProps {
  elementID: string,
  sitekey: string,
  callback: (response: string) => void
}

export interface RecaptchaState {
  ready: boolean,
  readyCheck: any,
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
        this.props.elementID,
        {
          sitekey: this.props.sitekey,
          callback: this.props.callback,
        }
      );
    }
  }

  private checkReady() {
    if (grecaptcha != undefined) {
      this.state.ready = true;
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
      <div id={this.props.elementID} data-sitekey={this.props.sitekey}></div>
    );
  }
}

