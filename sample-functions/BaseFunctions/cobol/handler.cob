       IDENTIFICATION DIVISION.
       PROGRAM-ID. APP.
      *> Example based upon http://stackoverflow.com/q/938760/1420197
      *> More on COBOL @ https://www.ibm.com/support/knowledgecenter/en/SS6SG3_3.4.0/com.ibm.entcobol.doc_3.4/tpbeg15.htm
       ENVIRONMENT DIVISION.
       INPUT-OUTPUT SECTION.
       FILE-CONTROL.
       SELECT SYSIN ASSIGN TO KEYBOARD ORGANIZATION LINE SEQUENTIAL.

       DATA DIVISION.
       FILE SECTION.
       FD SYSIN.
       01 ln PIC X(64).
          88 EOF VALUE HIGH-VALUES.

       WORKING-STORAGE SECTION.
       PROCEDURE DIVISION.
 
       DISPLAY "Request data: "      
       DISPLAY "------------"

       OPEN INPUT SYSIN
       READ SYSIN
       AT END SET EOF TO TRUE
       END-READ
       PERFORM UNTIL EOF


       DISPLAY ln

       READ SYSIN
       AT END SET EOF TO TRUE
       END-READ
       END-PERFORM
       CLOSE SYSIN

       DISPLAY "------------"
       STOP RUN.
