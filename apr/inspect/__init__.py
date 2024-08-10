'''
APR Inspection
'''
import apr.inspect.gui


def entry_point():
    application = apr.inspect.gui.RootWindow()
    application.mainloop()
