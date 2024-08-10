'''
Root Inspection Window
'''
# Python
import logging
import tkinter
import tkinter.ttk

# APR
import apr.inspect.menu
import apr.inspect.filenav
import apr.inspect.review


class RootWindow(tkinter.Tk):
    '''
    Main Tk window/container for application and runtime.
    '''
    def __init__(self, *args, **kwargs):
        super().__init__(*args, **kwargs)

        # Primary application frames
        self.mainframes = {
                'review': apr.inspect.review.VideoReview,
                'filenav': apr.inspect.filenav.FileSelection,
                }
        self.winfo_toplevel().target_file = tkinter.StringVar(self)
        self.winfo_toplevel().target_dir = tkinter.StringVar(self)
        self.winfo_toplevel().tempdir = None

        # Window properties
        self.title('APR Inspector')
        self.minsize(800, 550)

        # Window style
        self.style = tkinter.ttk.Style(self)
        try:
            self.style.theme_use('plastik')
        except tkinter.TclError:
            self.style.theme_use('alt')
        self.load_styles()

        # Main menu
        self.menu = apr.inspect.menu.MainMenu(self)
        self.config(menu=self.menu)

        # Status bar
        self.status = tkinter.ttk.Label(
                self, relief='sunken', anchor='w')
        self.status.grid(row=1, column=0, sticky='sew')

        # Primary viewport
        self.mainframe = tkinter.ttk.Frame(self)
        self.mainframe.body = None
        self.mainframe.grid(row=0, column=0, sticky='nsew')

        # Initial viewport frame
        self.set_mainframe('filenav')

        # Fill entire area
        self.grid_rowconfigure(0, weight=1)
        self.grid_columnconfigure(0, weight=1)

    def set_mainframe(self, framename):
        '''
        Reset main window with new mainframe.
        '''
        if self.mainframe.body:
            logging.debug('Purging old frame %s', self.mainframe.body)
            self.mainframe.body.grid_forget()
        logging.debug('Loading new frame %s', framename)
        self.mainframe.body = self.mainframes[framename](self)
        self.mainframe.body.grid(row=0, column=0, sticky='nsew')

        # Toggle menu entries
        if framename == 'review':
            self.menu.file.entryconfig('Select File', state='normal')
            # Hijack up button
            self.bind('<Up>', self.mainframe.body.play_frame)
            # TODO: Swap with <Enter>
        elif framename == 'filenav':
            self.menu.file.entryconfig('Select File', state='disabled')
            # Release up button
            self.unbind('<Up>')

    def load_styles(self):
        '''
        Load a pre-defined set of styles.
        '''
        # Vertical Scrollbar / removes all three arrows
        self.style.layout('arrowless.Vertical.Scrollbar', [
            ('Vertical.Scrollbar.trough', {
                'children': [('Vertical.Scrollbar.thumb', {})],
                })])

        # Horizontal Scrollbar / removes all three arrows
        self.style.layout('arrowless.Horizontal.Scrollbar', [
            ('Horizontal.Scrollbar.trough', {
                'children': [('Horizontal.Scrollbar.thumb', {})],
                })])
