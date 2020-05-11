import os
import requests
import discord
import json
import smtplib
import traceback
from random import randint
from email.mime.text import MIMEText

BOT_TOKEN = os.environ['BOT_TOKEN']
GIPHY_AUTH = os.environ['GIPHY_AUTH']
EMAIL_PASSWORD = os.environ['EMAIL_PASSWORD']
VER = '2.1.4'

msg_list = ['!help:      Shows this help message.\n',
            '!jeopardy:  Receive a category with 5 questions and answers. The ' +
                        'answers\n            are marked as spoilers and are not ' +
                        'revealed until you click\n            them XD\n',
            '!whisper:   Get a salty DM from SaltBot. This can be used as a ' +
                        'playground\n            for experiencing all of the ' +
                        'salty features.\n',
            '!hi:        Be greeted by SaltBot with a little added salt\n',
            '!goodnight: Hear a salty goodnight from SaltBot\n',
            '!gif:       Type "!gif" followed by keywords to get a cool gif ' +
                        'For\n            example: !gif dog\n',
            '!waifu:     Get a picture of a personal waifu that\'s different ' +
                        'each time\n',
            '!anime:     Get an anime recommendation just for you UwU\n',
            '!nut:       Receive a funny nut \'n go line']

with open('log.txt', 'w'):
    pass

client = discord.Client()

def help_fun(user=''):
    '''
        This function returns a help message thag gives a list of commands
    '''
    ret_msg = '```Good salty day to you ' + user + '! Here\'s a list of ' + \
          'commands that I understand:\n\n'
    for msg in msg_list:
        ret_msg+=msg

    ret_msg+='\n\nIf you have any further questions/concerns or if SaltBot ' + \
             'goes down, please\nhesitate to contact my developer: ' + \
             'HighSaltLevels. He\'s salty enough\nwithout your help and ' + \
             'doesn\'t write buggy code.\n\nCurrent Version: {}```'.format(VER)

    return ret_msg

def jeopardy(user=''):
    '''
        This function returns a 5 jeopardy questions and answers
    '''
    # Get a random set of questions
    rand = randint(0, 18417)
    resp = requests.get('http://jservice.io/api/category?id={}'.format(rand))

    # Verify status code
    if resp.status_code != 200:
        return'```I\'m Sorry. Something went wrong getting the questions```'

    # Convert to a json
    q_and_a = json.loads(resp.text)

    # Build and return the questions and answers
    msg = 'The Category is: "{}"\n\n'.format(q_and_a['title'])
    for i in range(5):
        msg+='Question {}: {}\nAnswer: ||{}||\n\n'.format(i+1, remove_crap(q_and_a['clues'][i]['question']), remove_crap(q_and_a['clues'][i]['answer']))

    return msg

def whisper(user=''):
    '''
        This function returns a hello message as a DM to the person who requested
    '''
    return '```Hello ' + user + '! You can talk to me here (Where no one ' + \
           'hear our mutual salt).```' 

def hi(user=''):
    '''
        This function reads greetings.txt and returns a random hello message
    '''

    # Read all of the greetings into memory
    with open('greetings.txt','r') as greets:
        hellos = [line for line in greets.readlines()]

        # Get a random number to index with and return the message
        msg = hellos[randint(0, len(hellos)-1)]
        return '```Hi {}! {}```'.format(user, msg)

def goodnight(user=''):
    '''
        This function reads goodnights.txt and returns a random goodnight message
    '''

    # Read all of the goodnights into memory
    with open('goodnights.txt','r') as nights:
        goodnights = [line for line in nights.readlines()]

        # Get a random number to index with and return the message
        msg = goodnights[randint(0, len(goodnights)-1)]
        return '```Goodnight {}! {}```'.format(user, msg)

def form_query(keywords):
    search_kw = ''
    for kw in keywords:
        search_kw+=(kw + '+')

    return f'http://api.giphy.com/v1/gifs/search?q={search_kw[:-1]}&api_key={GIPHY_AUTH}'

def send_all_gifs(keywords='whoops'):
    url = form_query(keywords)
    resp = requests.get(url)
    if resp.status_code != 200:
        return '```Sorry, I had trouble getting that gif :(```'

    json_resp = resp.json()
    
    num_gifs = len(json_resp['data'])
    if num_gifs == 0:
        return '```Sorry, there were no gifs of that query :(```'

    return [json_resp['data'][index]['bitly_gif_url'] for index in range(num_gifs)]

def gif(keywords='whoops', index=None):
    '''
        This function uses the giphy api to query and return a gif
    '''
    url = form_query(keywords)
    resp = requests.get(url)
    if resp.status_code != 200:
        return '```Sorry, I had trouble getting that gif :(```'

    json_resp = resp.json()

    num_gifs = len(json_resp['data'])
    if num_gifs == 0:
        return '```Sorry, there were no gifs of that query :(```'

    if index:
        try:
            index = int(index)
        except ValueError:
            return '```You have to specify a number between 0 and 24 if you want ' + \
                   'query by index!```'

        if index >= num_gifs:
            return '```Sorry, I didn\'t have enough gifs to get to that index```'

        if index < 0 or index > 24:
            return '```The index must be between 0 and 24```'

    else:
        index = randint(0, num_gifs - 1)

    return json_resp['data'][index]['bitly_gif_url']

def waifu():
    rand = randint(0, 99999)
    url = f'https://www.thiswaifudoesnotexist.net/example-{rand}.jpg'
    for _ in range(5):
        resp = requests.get(url, stream=True)
        if resp.status_code == 200:
            with open('temp.jpg', 'wb') as fw:
                fw.write(resp.content)
            return 'temp.jpg'
        else:
            log('admin', f'Got HTTP Status {resp.status_code} with content -> {resp.text}')
    return None

def anime():
    headers = {'User-Agent': 'Mozilla/5.0 (X11; Linux i586; rv:63.0) Gecko/20100101 Firefox/63.0.'}
    resp = requests.get('https://anidb.net/anime/random', headers=headers)
    if resp.status_code != 200:
        return '```Sowwy. Couldn\'t connect to the internet to get an anime recommendation :(```'
    data = resp.text
    title_idx = data.find('<title>')
    title_idx+=1
    title = ''
    for char in data[title_idx:]:
        if char == '<':
            break
        else:
            title+=char
    title = title[6:-15]

    description_idx = data.find('content=')
    description_idx+=8
    description = ''
    for char in data[description_idx:]:
        if char == '/':
            break
        else:
            description+=char
    return '```Here\'s an anime for you:\n\nTitle:\n{}\n\nDescription:\n{}```'.format(title, description)

def nut(user=''):
    with open('nut.txt') as fread:
        lines = [line for line in fread.readlines()]

    rand = randint(0, len(lines)-1)
    return f'```Remember {user}, don\'t {lines[rand]}```'

def remove_crap(orig_text):
    return orig_text.replace('<i>','').replace('</i>','').replace('<b>','').replace('</b>','').replace('\\',' ')

def log(author, msg):
    print(f'{author}: {msg}')
    with open('log.txt', 'a') as f:
        f.write(f'{author}: {msg}\n')

cmd_dict = {'!help':      help_fun,
            '!jeopardy':  jeopardy,
            '!whisper':   whisper,
            '!hi':        hi,
            '!goodnight': goodnight,
            '!gif':       gif,
            '!gifall':    send_all_gifs,
            '!waifu':     waifu,
            '!anime':     anime,
            '!nut':       nut,
            '!h':         help_fun,
            '!j':         jeopardy,
            '!pm':         whisper,
            '!i':         hi,
            '!n':         goodnight,
            '!g':         gif,
            '!gall':      send_all_gifs,
            '!w':         waifu,
            '!a':         anime,
            '!u':         nut}

@client.event
async def on_message(msg):
    # Only do something if command starts with ! or bot is not sending message
    if msg.author != client.user and msg.content.startswith('!'):
        # Get the command
        cmd = msg.content.split(' ')[0]

        # Notify user/channel if that command does not exist
        if cmd not in cmd_dict.keys():
            ret_msg = ("```Hello. I'm sorry I don't understand " + cmd + ". " 
                  " Please type \"!help\" to see a list of available commands"
                  "\n```")
            log(msg.author, 'invalid command {}'.format(cmd))
            await client.send_message(msg.channel, ret_msg)
        
        # Otherwise, call its assiciated function to get the string to send
        else:
            # Parse the author's username
            author = str(msg.author).split('#')[0]

            # If it is a whisper request...
            if cmd in ('!pm', '!whisper'):
                log(msg.author, 'dm sent')
                await client.send_message(msg.author, cmd_dict[cmd](author))

            # Or if it is a gif request...
            elif cmd.startswith('!g'):
                # Get the keywords and grab the index if specified
                keywords = str(msg.content).split(' ')
                keywords.remove(cmd)
                if not keywords:
                    log(msg.author, 'no gif to query')
                    message = '```You didn\'t give me a keyword to query :(```'
                    await client.send_message(msg.channel, message)
                else:
                    log(msg.author, 'sending gif - "{}"'.format(keywords))
                    idx = None
                    if '-i' in keywords:
                        i = 0
                        for i in range(len(keywords)):
                            if keywords[len(keywords)-1] == '-i':
                                message = '```Sorry, the last keyword cannot be "-i"```'
                                await client.send_message(msg.channel, message)
                                return
                            if keywords[i] == '-i':
                                idx = keywords[i+1]
                                keywords.remove('-i')
                                keywords.remove(idx)
                                break
                    elif '-a' in keywords:
                        if not isinstance(msg.channel, discord.channel.PrivateChannel):
                            err = '```You can only use -a in a DM!```'
                            await client.send_message(msg.channel, err)
                            return
                        else:
                            keywords.remove('-a')
                            gifs = cmd_dict[cmd+'all'](keywords)
                            for gif in gifs:
                                await client.send_message(msg.channel, gif)
                            return

                    await client.send_message(msg.channel, cmd_dict[cmd](keywords, idx))

            # Or if it is a waifu request...
            elif cmd.startswith('!w'):
                waifu = cmd_dict[cmd]()
                # Send the file instead
                if waifu:
                    log(msg.author, 'sending waifu')
                    await client.send_file(msg.channel, waifu)
                else:
                    error = '```Sorry. Couldn\'t grab that waifu picture. The internet must ' + \
                            ' broken again :(```'
                    log(msg.author, 'failed to send waifu')
                    await client.send_message(msg.channel, error)

            # Or if it as an anime request...
            elif cmd.startswith('!a'):
                message = '```Searching my database for the perfect anime UwU. Plz be patient...```'
                await client.send_message(msg.channel, message)
                log(msg.author, 'sending anime')
                await client.send_message(msg.channel, cmd_dict[cmd]())

            # Otherwise, it is a regular command
            else:
                log(msg.author, 'sending {}'.format(cmd))
                await client.send_message(msg.channel, cmd_dict[cmd](author))

@client.event
async def on_ready():
    print('Logged in as')
    print(client.user.name)
    print(client.user.id)
    print('------')
    await client.change_presence(game=discord.Game(name="The Salt Shaker"))

try:
    client.run(BOT_TOKEN)
except Exception as error:
    error_msg = traceback.format_exc()
    print(error_msg) 
    s=smtplib.SMTP('smtp.gmail.com', 587)
    s.ehlo()
    s.starttls()
    s.login('swdrummer13', EMAIL_PASSWORD)
    body = 'SaltBot crashed!\n' + error_msg
    msg = MIMEText(body)
    msg['Subject'] = 'SaltBot Crashed'
    msg['From'] = 'Me'
    msg['To'] = 'davidgreeson13@gmail.com'
    s.send_message(msg)
    s.quit()
    os.system('reboot')
