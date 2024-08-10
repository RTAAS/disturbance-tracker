'''
Main Menu
'''
# Python
import tkinter.ttk
import tkinter.filedialog


class MainMenu(tkinter.Menu):
    '''
    Application menu.
    '''
    def __init__(self, parent, *args, **kwargs):
        super().__init__(parent, *args, **kwargs)
        self.parent = parent

        # File
        self.file = tkinter.Menu(self, tearoff=0)
        self.add_cascade(label='File', menu=self.file)
        self.file.add_command(
                label='Select File', command=self.show_filenav)
        self.file.add_separator()
        self.file.add_command(
                label='Exit', command=self.master.destroy)

    def show_filenav(self):
        '''
        Display the package navigation pane.
        '''
        self.parent.set_mainframe('filenav')
