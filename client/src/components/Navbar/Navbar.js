import React, { Component } from 'react';
import * as AuthService from "../Auth/AuthService";
import { Redirect } from "react-router-dom";
import './Navbar.css';

class Navbar extends Component {
    state = { redirectToReferrer: false };
    logoutHandler = () => {
        AuthService.logout();
        this.setState({ redirectToReferrer: true });
    };
    render() {
        let { redirectToReferrer } = this.state;
        if (redirectToReferrer) return <Redirect to="/" />;
        return (
            <div className='nav-bar'>
                <button style={{ color: 'white' }} onClick={this.logoutHandler}>
                    Logout
                </button>
                
            </div>
        );
    }
}

export default Navbar;
