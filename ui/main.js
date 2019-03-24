#!/usr/bin/env node
/*
    Akmey is a web/ssh keyserver for SSH keys
    Copyright (C) 2019 Akmey contributors

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

const fs = require('fs');
const ini = require('ini');
const ora = require('ora');
const chalk = require('chalk');
const inquirer = require('inquirer');
const axios = require('axios');
const util = require('util');
const key = process.argv[2];
const mailregex = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;
const config = ini.parse(fs.readFileSync('config.ini', 'utf-8'));

axios.defaults.baseURL = config.clientlink.url;

function reportErr(err) {
    console.log(chalk.red('Ow...'));
    console.log(chalk.yellow(err.response.data.message));
    Object.entries(err.response.data.errors).forEach(val => {
        console.log('  ' + chalk.yellow(val[1][0]));
    });
    start();
}

function start() {
    axios.get('/api/user').then(userdata => {
        userdata = userdata.data;
        console.log(chalk.green('Hello ' + userdata.name + '!'));
        inquirer.prompt([
            {
                type: 'list',
                name: 'action',
                message: 'Choose an action down below :',
                choices: [
                    {
                        name: 'Add my key to Akmey',
                        value: 'add'
                    },
                    {
                        name: 'Edit my keys',
                        value: 'edit'
                    },
                    {
                        name: 'Remove a key from Akmey',
                        value: 'remove'
                    },
                    new inquirer.Separator(),
                    {
                        name: 'About Akmey',
                        value: 'about'
                    },
                    {
                        name: 'Quit',
                        value: 'quit'
                    }
                ]
            }
        ]).then(choice => {
            choice = choice.action;
            switch (choice) {
                case 'add':
                    var spin = ora('Adding your key').start();
                    axios.post('/api/keys/add', {key}).then(data => {
                        spin.succeed('Your key was successfully added!');
                        start();
                    }).catch(err => {
                        spin.fail();
                        reportErr(err);
                    });
                    break;
                
                case 'edit':
                    if (userdata.keys.length == 0) {
                        console.log(chalk.red('You don\'t have any keys!'));
                    } else {
                        var choices = [];
                        userdata.keys.forEach(obj => {
                            if (obj.comment) {
                                choices.push({
                                    name: obj.comment + ' (' + obj.key.substring(0, 90) + '...)',
                                    value: obj
                                });
                            } else {
                                choices.push({
                                    name: obj.key.substring(0, 90) + '...',
                                    value: obj
                                });
                            }
                        });
                        inquirer.prompt([
                            {
                                type: 'list',
                                name: 'key',
                                message: 'What key you want to edit?',
                                choices
                            },
                            {
                                type: 'input',
                                name: 'comment',
                                message: 'Name of the key?'
                            }
                        ]).then(ans => {
                            var spin = ora('Editing key...').start();
                            axios.put('/api/keys/'+ans.key.id, {comment: ans.comment}).then(res => {
                                spin.succeed('Key is now named ' + res.data.key.comment);
                                start();
                            }).catch(err => {
                                spin.fail('Cannot do that.');
                                reportErr();
                            });
                        });
                    }
                    break;

                case 'remove':
                    if (userdata.keys.length == 0) {
                        console.log(chalk.red('You don\'t have any keys!'));
                    } else {
                        var choices = [];
                        userdata.keys.forEach(obj => {
                            if (obj.comment) {
                                choices.push({
                                    name: obj.comment + ' (' + obj.key.substring(0, 90) + '...)',
                                    value: obj
                                });
                            } else {
                                choices.push({
                                    name: obj.key.substring(0, 90) + '...',
                                    value: obj
                                });
                            }
                        });
                        inquirer.prompt([
                            {
                                type: 'list',
                                name: 'key',
                                message: 'What key you want to remove?',
                                choices
                            }
                        ]).then(ans => {
                            var spin = ora('Deleting key...').start();
                            axios.delete('/api/keys/'+ans.key.id).then(res => {
                                spin.succeed('Deleted!');
                                start();
                            }).catch(err => {
                                spin.fail('Cannot do that.');
                                reportErr();
                            });
                        });
                    }
                    break;

                case 'about':
                    console.log('Akmey is like keyservers, but not for GPG keys, for SSH ones. It is bundled with clients for your servers. No need to remember your keys.');
                    start();
                    break;

                case 'quit':
                    console.log(chalk.green('Buh-bye!'));
                    process.exit(0);
                    break;

                default:
                    console.log(chalk.red('wtf?'));
                    process.exit(1);
                    break;
            }
        });
    }).catch(err => {
        console.log(chalk.red('Something weird happened, please contact the server admin. (Cannot retreive userdata)'));
    });
}

function init() {
    console.log(key);
    inquirer.prompt([
        {
            type: 'confirm',
            name: 'keyconf',
            message: 'Is that your key?'
        }
    ]).then(ans => {
        if (!ans.keyconf) {
            console.log(chalk.yellow('Check your SSH client configuration, and retry.'));
        } else {
            console.log(chalk.green('Welcome to Akmey! Before we start, you need to login.'));
            inquirer.prompt([
                {
                    type: 'input',
                    message: 'E-Mail',
                    name: 'username',
                    validate: val => {
                        if (mailregex.test(val)) {
                            return true;
                        } else {
                            return 'Please enter a valid e-mail address.'
                        }
                    }
                },
                {
                    type: 'password',
                    message: 'Password',
                    name: 'password',
                    mask: '*',
                    validate: val => {
                        if (val.length < 8) {
                            return 'Password must be 8 characters or more'
                        } else {
                            return true;
                        }
                    }
                }
            ]).then(authdata => {
                var spin = ora('Logging in...').start();
                Object.assign(authdata, {grant_type: 'password', client_id: config.clientlink.clientid, client_secret: config.clientlink.clientsecret, scope: 'keys'});
                axios.post('/oauth/token', authdata).then(response => {
                    axios.defaults.headers.common['Authorization'] = 'Bearer ' + response.data.access_token;
                    spin.succeed('Logged in!');
                    start();
                }).catch(err => {
                    spin.fail();
                    console.log(chalk.red('Oops! ' + err.response.data.message));
                    init();
                });
            });
        }    
    });
}

init();