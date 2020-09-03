import React, { Fragment, Component } from "react";
import * as AuthService from "./AuthService";
import { Redirect } from "react-router-dom";
import "./main.css";

const validUser = "admin"
const validPass = "123456"

class NotifyLoginComponent extends Component {
	
	constructor(props) {
		super(props)
		this.state = {
			isShow: 0,
			username: null,
			password: null,
		}
	}
	state = { redirectToReferrer: false };

	handleShow(e) {
		this.setState({
			isShow: this.state.isShow == 0 ? 1 : 0,
		});
	}
	handleInputUser(e) {
		this.setState({
			username: e.target.value,
		});
	}
	handleInputPass(e) {
		this.setState({
			password: e.target.value,
		});
	}
	loginHandler = () => {
		var success = 0;
		if (this.state.username == validUser && this.state.password == validPass)
			success = 1
		if (success == 1) {
			console.log("Came to login handler");
			AuthService.login();
			this.setState({ redirectToReferrer: true });
			this.setState({
				isSuccess: 1,
			});
		} else {
			this.setState({
				isSuccess: 0,
			});
		}
        
    };
    render() {
        let { from } = this.props.location.state || { from: { pathname: "/" } };
        let { redirectToReferrer } = this.state;
        if (redirectToReferrer) return <Redirect to={from} />;

        return (
			<body>

				<div class="limiter login">
					<div class="container-login100">
						<div class="wrap-login100">
							
							<div class="login100-form validate-form">
								<span class="login100-form-title p-b-26">
									Gosysmon
					</span>

								<div class={this.state.isSuccess == 0 ? "login-fail" : "login-fail-noactive"}>Invalid username or password!</div>
								<div class={this.state.isSuccess == 0 ? "spacing-31" : "spacing-70"}></div>

								<div class="wrap-input100 validate-input">
									<p>Username: </p>
									<input type="text" name="user" onInput={this.handleInputUser.bind(this)}></input>
							
					</div>

									<div class="wrap-input100 validate-input" data-validate="Enter password">
									<p>Password: </p>
									<span class="btn-pass" onClick={this.handleShow.bind(this)}>
										<i class={this.state.isShow == 0 ? "fa fa-eye" : "fa fa-eye-slash"}></i>
									</span>

									<input type={this.state.isShow == 0 ? "password" : "text"} name="pass" onInput={this.handleInputPass.bind(this)}></input>
									
							
					</div>

										<div class="container-login100-form-btn">
											<div class="wrap-login100-form-btn">
												<div class="login100-form-bgbtn"></div>
												<button onClick={this.loginHandler} class="login100-form-btn1">
													Login
							</button>
											</div>
										</div>

										<div class="text-center p-t-115">
											
										</div>
				</div>
								</div>
		</div>
						</div>
</body>
        );
    }
}
export default NotifyLoginComponent;

